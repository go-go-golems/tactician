package goals

import (
	"context"
	"regexp"
	"sort"
	"strings"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/go-go-golems/tactician/pkg/commands/sections"
	"github.com/go-go-golems/tactician/pkg/db"
	"github.com/go-go-golems/tactician/pkg/store"
	"github.com/pkg/errors"
)

type GoalsCommand struct {
	*cmds.CommandDefinition
}

type GoalsSettings struct {
	Mermaid bool `glazed.parameter:"mermaid"`
}

func NewGoalsCommand() (*GoalsCommand, error) {
	glazedSection, err := schema.NewGlazedSchema()
	if err != nil {
		return nil, err
	}

	tacticianSection, err := sections.NewTacticianSection()
	if err != nil {
		return nil, err
	}

	defaultSection, err := schema.NewSection(
		schema.DefaultSlug,
		"Default",
		schema.WithDescription("Default parameters"),
		schema.WithFields(
			fields.New("mermaid", fields.TypeBool,
				fields.WithHelp("Output as Mermaid diagram"),
				fields.WithDefault(false),
			),
		),
	)
	if err != nil {
		return nil, err
	}

	s := schema.NewSchema(schema.WithSections(glazedSection, tacticianSection, defaultSection))

	cmdDef := cmds.NewCommandDefinition(
		"goals",
		cmds.WithShort("List all open (incomplete) goals"),
		cmds.WithSchema(s),
	)

	return &GoalsCommand{CommandDefinition: cmdDef}, nil
}

var _ cmds.GlazeCommand = &GoalsCommand{}

func (c *GoalsCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	vals *values.Values,
	gp middlewares.Processor,
) error {
	tSettings := &sections.TacticianSettings{}
	if err := values.DecodeSectionInto(vals, sections.TacticianSlug, tSettings); err != nil {
		return errors.Wrap(err, "decode tactician settings")
	}

	settings := &GoalsSettings{}
	if err := values.DecodeSectionInto(vals, schema.DefaultSlug, settings); err != nil {
		return errors.Wrap(err, "decode goals settings")
	}

	st, err := store.Load(ctx, tSettings.Dir)
	if err != nil {
		return err
	}
	defer func() { _ = st.Close() }()

	allNodes, err := st.Project.GetAllNodes(ctx)
	if err != nil {
		return err
	}

	var pending []*db.Node
	for _, n := range allNodes {
		if n.Status == "pending" {
			pending = append(pending, n)
		}
	}

	if settings.Mermaid {
		mermaid, err := buildMermaidGoals(ctx, st, pending)
		if err != nil {
			return err
		}
		return gp.AddRow(ctx, types.NewRow(types.MRP("mermaid", mermaid)))
	}

	if len(pending) == 0 {
		return gp.AddRow(ctx, types.NewRow(types.MRP("message", "All goals complete!")))
	}

	type nodeInfo struct {
		n      *db.Node
		status string
		deps   []string
		blocks []string
	}
	var infos []nodeInfo
	for _, n := range pending {
		status, err := computeNodeStatus(ctx, st, n.ID)
		if err != nil {
			return err
		}
		depsNodes, err := st.Project.GetDependencies(ctx, n.ID)
		if err != nil {
			return err
		}
		var deps []string
		for _, d := range depsNodes {
			deps = append(deps, d.ID)
		}
		blocksNodes, err := st.Project.GetBlockedBy(ctx, n.ID)
		if err != nil {
			return err
		}
		var blocks []string
		for _, b := range blocksNodes {
			blocks = append(blocks, b.ID)
		}

		infos = append(infos, nodeInfo{n: n, status: status, deps: deps, blocks: blocks})
	}

	// Ready first.
	sort.SliceStable(infos, func(i, j int) bool {
		if infos[i].status == "ready" && infos[j].status != "ready" {
			return true
		}
		if infos[i].status != "ready" && infos[j].status == "ready" {
			return false
		}
		return infos[i].n.ID < infos[j].n.ID
	})

	for _, info := range infos {
		row := types.NewRow(
			types.MRP("id", info.n.ID),
			types.MRP("output", info.n.Output),
			types.MRP("status", info.status),
			types.MRP("dependencies", strings.Join(info.deps, ",")),
			types.MRP("blocks", strings.Join(info.blocks, ",")),
			types.MRP("parent_tactic", info.n.ParentTactic),
		)
		if err := gp.AddRow(ctx, row); err != nil {
			return err
		}
	}

	return nil
}

func computeNodeStatus(ctx context.Context, st *store.State, nodeID string) (string, error) {
	n, err := st.Project.GetNode(ctx, nodeID)
	if err != nil {
		return "", err
	}
	if n == nil {
		return "", errors.Errorf("node not found: %s", nodeID)
	}
	if n.Status == "complete" {
		return "complete", nil
	}
	deps, err := st.Project.GetDependencies(ctx, nodeID)
	if err != nil {
		return "", err
	}
	if len(deps) == 0 {
		return "ready", nil
	}
	for _, d := range deps {
		if d.Status != "complete" {
			return "blocked", nil
		}
	}
	return "ready", nil
}

var mermaidSanitizeRe = regexp.MustCompile(`[^a-zA-Z0-9_]`)

func mermaidID(id string) string {
	return mermaidSanitizeRe.ReplaceAllString(id, "_")
}

func mermaidLabel(id, output, typ, status string) string {
	parts := []string{}
	id = strings.TrimSpace(id)
	output = strings.TrimSpace(output)
	typ = strings.TrimSpace(typ)
	status = strings.TrimSpace(status)

	if id != "" {
		parts = append(parts, id)
	}
	if output != "" && output != id {
		parts = append(parts, output)
	}
	if typ != "" {
		parts = append(parts, typ)
	}
	if status != "" {
		parts = append(parts, "["+strings.ToUpper(status)+"]")
	}
	if len(parts) == 0 {
		return ""
	}
	return strings.Join(parts, "<br/>")
}

func buildMermaidGoals(ctx context.Context, st *store.State, pending []*db.Node) (string, error) {
	sb := strings.Builder{}
	sb.WriteString("graph TD\n")
	if len(pending) == 0 {
		sb.WriteString("  empty[\"All goals complete!\"]\n")
		return sb.String(), nil
	}

	// Nodes
	for _, n := range pending {
		status, err := computeNodeStatus(ctx, st, n.ID)
		if err != nil {
			return "", err
		}
		sb.WriteString("  ")
		sb.WriteString(mermaidID(n.ID))
		sb.WriteString("[\"")
		label := mermaidLabel(n.ID, n.Output, n.Type, status)
		sb.WriteString(strings.ReplaceAll(label, "\"", "\\\""))
		sb.WriteString("\"]\n")
	}

	// Edges from deps.
	for _, n := range pending {
		deps, err := st.Project.GetDependencies(ctx, n.ID)
		if err != nil {
			return "", err
		}
		for _, d := range deps {
			sb.WriteString("  ")
			sb.WriteString(mermaidID(d.ID))
			sb.WriteString(" --> ")
			sb.WriteString(mermaidID(n.ID))
			sb.WriteString("\n")
		}
	}

	return sb.String(), nil
}
