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

type NodeAddCommand struct {
	*cmds.CommandDefinition
}

func NewNodeAddCommand() (*NodeAddCommand, error) {
	projectSection, err := sections.NewProjectSection()
	if err != nil {
		return nil, err
	}

	defaultSection, err := schema.NewSection(
		schema.DefaultSlug,
		"Default",
		schema.WithDescription("Default parameters"),
		schema.WithArguments(
			fields.New("node-id", fields.TypeString,
				fields.WithHelp("Node ID"),
				fields.WithRequired(true),
			),
			fields.New("output", fields.TypeString,
				fields.WithHelp("Node output artifact"),
				fields.WithRequired(true),
			),
		),
		schema.WithFields(
			fields.New("type", fields.TypeString,
				fields.WithHelp("Node type"),
				fields.WithDefault("project_artifact"),
			),
			fields.New("status", fields.TypeChoice,
				fields.WithHelp("Initial status"),
				fields.WithChoices("pending", "complete"),
				fields.WithDefault("pending"),
			),
		),
	)
	if err != nil {
		return nil, err
	}

	s := schema.NewSchema(schema.WithSections(projectSection, defaultSection))

	cmdDef := cmds.NewCommandDefinition(
		"add",
		cmds.WithShort("Add a new node"),
		cmds.WithSchema(s),
	)

	return &NodeAddCommand{CommandDefinition: cmdDef}, nil
}

var _ cmds.BareCommand = &NodeAddCommand{}

func (c *NodeAddCommand) Run(ctx context.Context, vals *values.Values) error {
	_ = ctx
	_ = vals
	return errors.New("node add: not implemented yet")
}
