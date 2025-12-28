package history

import (
	"context"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/pkg/errors"
)

type HistoryCommand struct {
	*cmds.CommandDefinition
}

func NewHistoryCommand() (*HistoryCommand, error) {
	glazedSection, err := schema.NewGlazedSchema()
	if err != nil {
		return nil, err
	}

	defaultSection, err := schema.NewSection(
		schema.DefaultSlug,
		"Default",
		schema.WithDescription("Default parameters"),
		schema.WithFields(
			fields.New("limit", fields.TypeInteger,
				fields.WithHelp("Limit number of entries"),
				fields.WithShortFlag("l"),
			),
			fields.New("since", fields.TypeString,
				fields.WithHelp("Show actions since (e.g., 1h, 2d, 30m)"),
				fields.WithShortFlag("s"),
			),
			fields.New("summary", fields.TypeBool,
				fields.WithHelp("Show session summary instead of detailed log"),
				fields.WithDefault(false),
			),
		),
	)
	if err != nil {
		return nil, err
	}

	s := schema.NewSchema(schema.WithSections(glazedSection, defaultSection))

	cmdDef := cmds.NewCommandDefinition(
		"history",
		cmds.WithShort("View action history and session summary"),
		cmds.WithSchema(s),
	)

	return &HistoryCommand{CommandDefinition: cmdDef}, nil
}

var _ cmds.GlazeCommand = &HistoryCommand{}

func (c *HistoryCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	vals *values.Values,
	gp middlewares.Processor,
) error {
	_ = ctx
	_ = vals
	_ = gp
	return errors.New("history: not implemented yet")
}
