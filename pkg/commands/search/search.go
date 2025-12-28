package search

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

type SearchCommand struct {
	*cmds.CommandDefinition
}

func NewSearchCommand() (*SearchCommand, error) {
	glazedSection, err := schema.NewGlazedSchema()
	if err != nil {
		return nil, err
	}

	projectSection, err := sections.NewProjectSection()
	if err != nil {
		return nil, err
	}

	defaultSection, err := schema.NewSection(
		schema.DefaultSlug,
		"Default",
		schema.WithDescription("Default parameters"),
		schema.WithArguments(
			fields.New("query", fields.TypeString,
				fields.WithHelp("Search query (keywords)"),
			),
		),
		schema.WithFields(
			fields.New("ready", fields.TypeBool,
				fields.WithHelp("Show only ready tactics (all dependencies satisfied)"),
				fields.WithDefault(false),
			),
			fields.New("type", fields.TypeString,
				fields.WithHelp("Filter by tactic type"),
			),
			fields.New("tags", fields.TypeString,
				fields.WithHelp("Filter by tags (comma-separated)"),
			),
			fields.New("goals", fields.TypeString,
				fields.WithHelp("Align with specific goal nodes (comma-separated)"),
			),
			fields.New("llm-rerank", fields.TypeBool,
				fields.WithHelp("Use LLM to semantically rerank results"),
				fields.WithDefault(false),
			),
			fields.New("limit", fields.TypeInteger,
				fields.WithHelp("Limit number of results"),
				fields.WithDefault(20),
				fields.WithShortFlag("l"),
			),
			fields.New("verbose", fields.TypeBool,
				fields.WithHelp("Show detailed scoring information"),
				fields.WithDefault(false),
				fields.WithShortFlag("v"),
			),
		),
	)
	if err != nil {
		return nil, err
	}

	s := schema.NewSchema(schema.WithSections(glazedSection, projectSection, defaultSection))

	cmdDef := cmds.NewCommandDefinition(
		"search",
		cmds.WithShort("Search for applicable tactics"),
		cmds.WithSchema(s),
	)

	return &SearchCommand{CommandDefinition: cmdDef}, nil
}

var _ cmds.GlazeCommand = &SearchCommand{}

func (c *SearchCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	vals *values.Values,
	gp middlewares.Processor,
) error {
	_ = ctx
	_ = vals
	_ = gp
	return errors.New("search: not implemented yet")
}
