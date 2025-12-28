package node

import (
	"context"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/go-go-golems/tactician/pkg/commands/sections"
	"github.com/pkg/errors"
)

type NodeDeleteCommand struct {
	*cmds.CommandDefinition
}

func NewNodeDeleteCommand() (*NodeDeleteCommand, error) {
	projectSection, err := sections.NewProjectSection()
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

	s := schema.NewSchema(schema.WithSections(projectSection, defaultSection))

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
	_ = ctx
	_ = vals
	return errors.New("node delete: not implemented yet")
}
