package graph

import (
	"context"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/tactician/pkg/commands/sections"
	"github.com/pkg/errors"
)

type GraphCommand struct {
	*cmds.CommandDefinition
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
	_ = ctx
	_ = vals
	_ = gp
	return errors.New("graph: not implemented yet")
}
