package node

import (
	"context"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/go-go-golems/tactician/pkg/commands/sections"
	"github.com/go-go-golems/tactician/pkg/store"
	"github.com/pkg/errors"
)

type NodeShowCommand struct {
	*cmds.CommandDefinition
}

type NodeShowSettings struct {
	NodeIDs []string `glazed.parameter:"node-ids"`
}

func NewNodeShowCommand() (*NodeShowCommand, error) {
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
			fields.New("node-ids", fields.TypeStringList,
				fields.WithHelp("Node ID(s) to show (supports multiple: node show id1 id2 id3)"),
				fields.WithRequired(true),
			),
		),
	)
	if err != nil {
		return nil, err
	}

	s := schema.NewSchema(
		schema.WithSections(glazedSection, tacticianSection, defaultSection),
	)

	cmdDef := cmds.NewCommandDefinition(
		"show",
		cmds.WithShort("Show details for one or more nodes"),
		cmds.WithLong("Show details for nodes. Supports batch operations: 'node show id1 id2 id3'"),
		cmds.WithSchema(s),
	)

	return &NodeShowCommand{CommandDefinition: cmdDef}, nil
}

var _ cmds.GlazeCommand = &NodeShowCommand{}

func (c *NodeShowCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	vals *values.Values,
	gp middlewares.Processor,
) error {
	tSettings := &sections.TacticianSettings{}
	if err := values.DecodeSectionInto(vals, sections.TacticianSlug, tSettings); err != nil {
		return errors.Wrap(err, "decode tactician settings")
	}

	settings := &NodeShowSettings{}
	if err := values.DecodeSectionInto(vals, schema.DefaultSlug, settings); err != nil {
		return errors.Wrap(err, "decode node show settings")
	}
	if len(settings.NodeIDs) == 0 {
		return errors.New("at least one node id is required")
	}

	st, err := store.Load(ctx, tSettings.Dir)
	if err != nil {
		return err
	}
	defer func() { _ = st.Close() }()

	for _, id := range settings.NodeIDs {
		n, err := st.Project.GetNode(ctx, id)
		if err != nil {
			return err
		}
		if n == nil {
			return errors.Errorf("node not found: %s", id)
		}

		row := types.NewRow(
			types.MRP("id", n.ID),
			types.MRP("type", n.Type),
			types.MRP("output", n.Output),
			types.MRP("status", n.Status),
			types.MRP("created_by", n.CreatedBy),
			types.MRP("parent_tactic", n.ParentTactic),
			types.MRP("introduced_as", n.IntroducedAs),
			types.MRP("created_at", n.CreatedAt),
			types.MRP("completed_at", n.CompletedAt),
		)
		if err := gp.AddRow(ctx, row); err != nil {
			return err
		}
	}

	return nil
}
