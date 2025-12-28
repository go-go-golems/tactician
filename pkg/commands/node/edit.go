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

type NodeEditCommand struct {
	*cmds.CommandDefinition
}

func NewNodeEditCommand() (*NodeEditCommand, error) {
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
				fields.WithHelp("Node ID(s) to edit (supports multiple: node edit id1 id2 --status complete)"),
				fields.WithRequired(true),
			),
		),
		schema.WithFields(
			fields.New("status", fields.TypeChoice,
				fields.WithHelp("Update status (applied to all specified nodes)"),
				fields.WithChoices("pending", "complete"),
				fields.WithRequired(true),
			),
		),
	)
	if err != nil {
		return nil, err
	}

	s := schema.NewSchema(schema.WithSections(projectSection, defaultSection))

	cmdDef := cmds.NewCommandDefinition(
		"edit",
		cmds.WithShort("Edit one or more nodes"),
		cmds.WithLong("Update status for nodes. Supports batch operations: 'node edit id1 id2 --status complete'"),
		cmds.WithSchema(s),
	)

	return &NodeEditCommand{CommandDefinition: cmdDef}, nil
}

var _ cmds.BareCommand = &NodeEditCommand{}

func (c *NodeEditCommand) Run(ctx context.Context, vals *values.Values) error {
	_ = ctx
	_ = vals
	return errors.New("node edit: not implemented yet")
}
