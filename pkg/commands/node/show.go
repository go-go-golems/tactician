package node

import (
	"context"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/pkg/errors"
)

type NodeShowCommand struct {
	*cmds.CommandDefinition
}

func NewNodeShowCommand() (*NodeShowCommand, error) {
	glazedSection, err := schema.NewGlazedSchema()
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
		schema.WithSections(glazedSection, defaultSection),
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
	_ = ctx
	_ = vals
	_ = gp
	return errors.New("node show: not implemented yet")
}
