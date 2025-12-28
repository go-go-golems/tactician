package graph

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

type GraphCommand struct {
	*cmds.CommandDefinition
}

type GraphSettings struct {
	GoalID  string `glazed.parameter:"goal-id"`
	Mermaid bool   `glazed.parameter:"mermaid"`
}

func NewGraphCommand() (*GraphCommand, error) {
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
		schema.WithArguments(
			fields.New("goal-id", fields.TypeString,
				fields.WithHelp("Start graph from specific goal node"),
			),
		),
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
		"graph",
		cmds.WithShort("Display the project dependency graph"),
		cmds.WithSchema(s),
	)

	return &GraphCommand{CommandDefinition: cmdDef}, nil
}

var _ cmds.GlazeCommand = &GraphCommand{}

func (c *GraphCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	vals *values.Values,
	gp middlewares.Processor,
) error {
	tSettings := &sections.TacticianSettings{}
	if err := values.DecodeSectionInto(vals, sections.TacticianSlug, tSettings); err != nil {
		return errors.Wrap(err, "decode tactician settings")
	}

	settings := &GraphSettings{}
	if err := values.DecodeSectionInto(vals, schema.DefaultSlug, settings); err != nil {
		return errors.Wrap(err, "decode graph settings")
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
	edges, err := st.Project.GetEdges(ctx)
	if err != nil {
		return err
	}
	meta, err := st.Project.GetProjectMeta(ctx)
	if err != nil {
		return err
	}

	if settings.Mermaid {
		mermaid, err := buildMermaidGraph(ctx, st, allNodes, edges)
		if err != nil {
			return err
		}
		row := types.NewRow(types.MRP("project", meta["name"]), types.MRP("mermaid", mermaid))
		return gp.AddRow(ctx, row)
	}

	if len(allNodes) == 0 {
		return gp.AddRow(ctx, types.NewRow(types.MRP("message", "No nodes in project yet.")))
	}

	byID := map[string]int{}
	for i, n := range allNodes {
		byID[n.ID] = i
	}

	rootID := strings.TrimSpace(settings.GoalID)
	if rootID == "" {
		rootID = strings.TrimSpace(meta["root_goal"])
	}
	if rootID == "" {
		// Find nodes with no incoming edges (roots).
		hasIncoming := map[string]bool{}
		for _, e := range edges {
			hasIncoming[e.TargetNodeID] = true
		}
		var roots []string
		for _, n := range allNodes {
			if !hasIncoming[n.ID] {
				roots = append(roots, n.ID)
			}
		}
		sort.Strings(roots)
		if len(roots) == 0 {
			return errors.New("no root node found (set root_goal in project meta or specify goal-id)")
		}
		rootID = roots[0]
	}

	if _, ok := byID[rootID]; !ok {
		return errors.Errorf("node not found: %s", rootID)
	}

	children := map[string][]string{}
	for _, e := range edges {
		children[e.SourceNodeID] = append(children[e.SourceNodeID], e.TargetNodeID)
	}
	for k := range children {
		sort.Strings(children[k])
	}

	visited := map[string]bool{}
	var walk func(id string, depth int) error
	walk = func(id string, depth int) error {
		if visited[id] {
			return nil
		}
		visited[id] = true

		n := allNodes[byID[id]]
		status, err := computeNodeStatus(ctx, st, n.ID)
		if err != nil {
			return err
		}

		row := types.NewRow(
			types.MRP("project", meta["name"]),
			types.MRP("root", rootID),
			types.MRP("depth", depth),
			types.MRP("id", n.ID),
			types.MRP("type", n.Type),
			types.MRP("output", n.Output),
			types.MRP("status", status),
		)
		if err := gp.AddRow(ctx, row); err != nil {
			return err
		}

		for _, child := range children[id] {
			if err := walk(child, depth+1); err != nil {
				return err
			}
		}
		return nil
	}

	return walk(rootID, 0)
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
	// Build a useful label without repeating id/output when they’re identical.
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

func buildMermaidGraph(ctx context.Context, st *store.State, nodes []*db.Node, edges []db.Edge) (string, error) {
	sb := strings.Builder{}
	sb.WriteString("graph TD\n")
	for _, n := range nodes {
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
	for _, e := range edges {
		sb.WriteString("  ")
		sb.WriteString(mermaidID(e.SourceNodeID))
		sb.WriteString(" --> ")
		sb.WriteString(mermaidID(e.TargetNodeID))
		sb.WriteString("\n")
	}
	return sb.String(), nil
}
