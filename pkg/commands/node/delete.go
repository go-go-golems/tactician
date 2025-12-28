package node

import (
	"context"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/go-go-golems/tactician/pkg/commands/sections"
	"github.com/go-go-golems/tactician/pkg/store"
	"github.com/pkg/errors"
)

type NodeDeleteCommand struct {
	*cmds.CommandDefinition
}

type NodeDeleteSettings struct {
	NodeIDs []string `glazed.parameter:"node-ids"`
	Force   bool     `glazed.parameter:"force"`
}

func NewNodeDeleteCommand() (*NodeDeleteCommand, error) {
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
				fields.WithHelp("Node ID(s) to delete (supports multiple: node delete id1 id2 --force)"),
				fields.WithRequired(true),
			),
		),
		schema.WithFields(
			fields.New("force", fields.TypeBool,
				fields.WithHelp("Force delete even if it blocks other nodes (applies to all specified nodes)"),
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
		"delete",
		cmds.WithShort("Delete one or more nodes"),
		cmds.WithLong("Delete nodes. Supports batch operations: 'node delete id1 id2 --force'"),
		cmds.WithSchema(s),
	)

	return &NodeDeleteCommand{CommandDefinition: cmdDef}, nil
}

var _ cmds.BareCommand = &NodeDeleteCommand{}

func (c *NodeDeleteCommand) Run(ctx context.Context, vals *values.Values) error {
	tSettings := &sections.TacticianSettings{}
	if err := values.DecodeSectionInto(vals, sections.TacticianSlug, tSettings); err != nil {
		return errors.Wrap(err, "decode tactician settings")
	}

	settings := &NodeDeleteSettings{}
	if err := values.DecodeSectionInto(vals, schema.DefaultSlug, settings); err != nil {
		return errors.Wrap(err, "decode node delete settings")
	}
	if len(settings.NodeIDs) == 0 {
		return errors.New("at least one node id is required")
	}

	st, err := store.Load(ctx, tSettings.Dir)
	if err != nil {
		return err
	}
	defer func() { _ = st.Close() }()

	// Validate
	for _, id := range settings.NodeIDs {
		n, err := st.Project.GetNode(ctx, id)
		if err != nil {
			return err
		}
		if n == nil {
			return errors.Errorf("node not found: %s", id)
		}
		if !settings.Force {
			blocks, err := st.Project.GetBlockedBy(ctx, id)
			if err != nil {
				return err
			}
			if len(blocks) > 0 {
				return errors.Errorf("cannot delete %s: it blocks %d node(s) (use --force)", id, len(blocks))
			}
		}
	}

	for _, id := range settings.NodeIDs {
		if err := st.Project.DeleteNode(ctx, id); err != nil {
			return err
		}
		details := "Deleted node: " + id
		nodeID := id
		if err := st.Project.LogAction(ctx, "node_deleted", &details, &nodeID, nil); err != nil {
			return err
		}
	}

	st.Dirty = true
	return st.Save(ctx)
}
