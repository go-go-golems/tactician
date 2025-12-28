package apply

import (
	"context"
	"encoding/json"
	"sort"
	"strings"
	"time"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/go-go-golems/tactician/pkg/commands/sections"
	"github.com/go-go-golems/tactician/pkg/db"
	"github.com/go-go-golems/tactician/pkg/store"
	"github.com/pkg/errors"
)

type ApplyCommand struct {
	*cmds.CommandDefinition
}

type ApplySettings struct {
	TacticID string `glazed.parameter:"tactic-id"`
	Yes      bool   `glazed.parameter:"yes"`
	Force    bool   `glazed.parameter:"force"`
}

func NewApplyCommand() (*ApplyCommand, error) {
	tacticianSection, err := sections.NewTacticianSection()
	if err != nil {
		return nil, err
	}

	defaultSection, err := schema.NewSection(
		schema.DefaultSlug,
		"Default",
		schema.WithDescription("Default parameters"),
		schema.WithArguments(
			fields.New("tactic-id", fields.TypeString,
				fields.WithHelp("Tactic ID to apply"),
				fields.WithRequired(true),
			),
		),
		schema.WithFields(
			fields.New("yes", fields.TypeBool,
				fields.WithHelp("Skip confirmation prompt"),
				fields.WithDefault(false),
				fields.WithShortFlag("y"),
			),
			fields.New("force", fields.TypeBool,
				fields.WithHelp("Apply even if dependencies are missing"),
				fields.WithDefault(false),
				fields.WithShortFlag("f"),
			),
		),
	)
	if err != nil {
		return nil, err
	}

	s := schema.NewSchema(schema.WithSections(tacticianSection, defaultSection))

	cmdDef := cmds.NewCommandDefinition(
		"apply",
		cmds.WithShort("Apply a tactic to create new nodes"),
		cmds.WithSchema(s),
	)

	return &ApplyCommand{CommandDefinition: cmdDef}, nil
}

var _ cmds.BareCommand = &ApplyCommand{}

func (c *ApplyCommand) Run(ctx context.Context, vals *values.Values) error {
	tSettings := &sections.TacticianSettings{}
	if err := values.DecodeSectionInto(vals, sections.TacticianSlug, tSettings); err != nil {
		return errors.Wrap(err, "decode tactician settings")
	}

	settings := &ApplySettings{}
	if err := values.DecodeSectionInto(vals, schema.DefaultSlug, settings); err != nil {
		return errors.Wrap(err, "decode apply settings")
	}

	if !settings.Yes {
		return errors.New("apply requires confirmation; re-run with --yes")
	}

	st, err := store.Load(ctx, tSettings.Dir)
	if err != nil {
		return err
	}
	defer func() { _ = st.Close() }()

	tactic, err := st.Tactics.GetTactic(ctx, settings.TacticID)
	if err != nil {
		return err
	}
	if tactic == nil {
		return errors.Errorf("tactic not found: %s", settings.TacticID)
	}

	allNodes, err := st.Project.GetAllNodes(ctx)
	if err != nil {
		return err
	}

	deps := checkDependencies(tactic, allNodes)
	if len(deps.Missing) > 0 && !settings.Force {
		return errors.Errorf("cannot apply tactic: missing required dependencies (%s); use --force", strings.Join(deps.Missing, ","))
	}

	now := time.Now().UTC()
	createdBy := "tactic:" + settings.TacticID

	// Determine nodes to create
	var nodesToCreate []*db.Node

	// Introduce premise nodes if needed
	for _, output := range deps.CanIntroduce {
		id := output
		introducedAs := "premise"
		node := &db.Node{
			ID:           id,
			Type:         "document",
			Output:       output,
			Status:       "pending",
			CreatedBy:    &createdBy,
			IntroducedAs: &introducedAs,
			CreatedAt:    now,
		}
		nodesToCreate = append(nodesToCreate, node)
	}

	// Subtask nodes or single output node
	if len(tactic.Subtasks) > 0 {
		for _, stt := range tactic.Subtasks {
			var data json.RawMessage
			if stt.Data != nil {
				b, err := json.Marshal(stt.Data)
				if err != nil {
					return errors.Wrap(err, "marshal subtask data")
				}
				data = b
			}
			parent := settings.TacticID
			node := &db.Node{
				ID:           stt.ID,
				Type:         stt.Type,
				Output:       stt.Output,
				Status:       "pending",
				CreatedBy:    &createdBy,
				ParentTactic: &parent,
				Data:         data,
				CreatedAt:    now,
			}
			nodesToCreate = append(nodesToCreate, node)
		}
	} else {
		var data json.RawMessage
		if tactic.Data != nil {
			b, err := json.Marshal(tactic.Data)
			if err != nil {
				return errors.Wrap(err, "marshal tactic data")
			}
			data = b
		}
		node := &db.Node{
			ID:        tactic.Output,
			Type:      tactic.Type,
			Output:    tactic.Output,
			Status:    "pending",
			CreatedBy: &createdBy,
			Data:      data,
			CreatedAt: now,
		}
		nodesToCreate = append(nodesToCreate, node)
	}

	// Validate none exist
	for _, n := range nodesToCreate {
		existing, err := st.Project.GetNode(ctx, n.ID)
		if err != nil {
			return err
		}
		if existing != nil {
			return errors.Errorf("node already exists: %s", n.ID)
		}
	}

	// Create nodes
	for _, n := range nodesToCreate {
		if err := st.Project.AddNode(ctx, n); err != nil {
			return err
		}
	}

	// Subtask dependency edges
	for _, stt := range tactic.Subtasks {
		for _, depID := range stt.DependsOn {
			if err := st.Project.AddEdge(ctx, depID, stt.ID); err != nil {
				return err
			}
		}
	}

	// Edges from match dependencies to created nodes.
	//
	// NOTE: "Satisfied" vs "missing" is about completion; the dependency edge should still exist
	// when the dependency node exists but is not complete (so downstream nodes show as blocked).
	for _, depOutput := range tactic.Match {
		source := findNodeByOutput(allNodes, depOutput)
		if source == nil {
			continue
		}
		for _, newNode := range nodesToCreate {
			if err := st.Project.AddEdge(ctx, source.ID, newNode.ID); err != nil {
				return err
			}
		}
	}

	// Edges from premise dependencies to created nodes.
	//
	// Premises that exist (whether complete or not) should create edges, just like match.
	// Premises that are being introduced (canIntroduce) are new nodes themselves, so skip those.
	canIntroduceSet := make(map[string]bool)
	for _, output := range deps.CanIntroduce {
		canIntroduceSet[output] = true
	}
	for _, depOutput := range tactic.Premises {
		// Skip if already handled as match dependency
		if contains(tactic.Match, depOutput) {
			continue
		}
		// Skip if being introduced (it's a new node itself)
		if canIntroduceSet[depOutput] {
			continue
		}
		// Create edges from existing premise nodes to new nodes
		source := findNodeByOutput(allNodes, depOutput)
		if source == nil {
			continue
		}
		for _, newNode := range nodesToCreate {
			if err := st.Project.AddEdge(ctx, source.ID, newNode.ID); err != nil {
				return err
			}
		}
	}

	details := "Applied tactic: " + settings.TacticID
	tid := settings.TacticID
	if err := st.Project.LogAction(ctx, "tactic_applied", &details, nil, &tid); err != nil {
		return err
	}

	st.Dirty = true
	return st.Save(ctx)
}

type depCheck struct {
	Satisfied    []string
	Missing      []string
	CanIntroduce []string
}

func checkDependencies(tactic *db.Tactic, allNodes []*db.Node) depCheck {
	completeOutputs := map[string]bool{}
	existingOutputs := map[string]bool{}
	for _, n := range allNodes {
		existingOutputs[n.Output] = true
		if n.Status == "complete" {
			completeOutputs[n.Output] = true
		}
	}

	satisfiedSet := map[string]bool{}
	missingSet := map[string]bool{}
	canIntroduceSet := map[string]bool{}

	for _, dep := range tactic.Match {
		if completeOutputs[dep] {
			satisfiedSet[dep] = true
		} else {
			missingSet[dep] = true
		}
	}

	for _, dep := range tactic.Premises {
		if contains(tactic.Match, dep) {
			continue
		}
		if completeOutputs[dep] {
			satisfiedSet[dep] = true
		} else if !existingOutputs[dep] {
			canIntroduceSet[dep] = true
		} else {
			missingSet[dep] = true
		}
	}

	return depCheck{
		Satisfied:    keys(satisfiedSet),
		Missing:      keys(missingSet),
		CanIntroduce: keys(canIntroduceSet),
	}
}

func keys(m map[string]bool) []string {
	var out []string
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func contains(list []string, s string) bool {
	for _, x := range list {
		if x == s {
			return true
		}
	}
	return false
}

func findNodeByOutput(nodes []*db.Node, output string) *db.Node {
	for _, n := range nodes {
		if n.Output == output && n.Status == "complete" {
			return n
		}
	}
	for _, n := range nodes {
		if n.Output == output {
			return n
		}
	}
	return nil
}
