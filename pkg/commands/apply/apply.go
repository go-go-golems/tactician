package apply

import (
	"context"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/pkg/errors"
)

type ApplyCommand struct {
	*cmds.CommandDefinition
}

func NewApplyCommand() (*ApplyCommand, error) {
	defaultSection, err := schema.NewSection(
		schema.DefaultSlug,
		"Default",
		schema.WithDescription("Default parameters"),
		schema.WithArguments(
			fields.New("tactic-id", fields.TypeString,
				fields.WithHelp("Tactic ID to apply"),
				fields.WithRequired(true),
			),
		),
		schema.WithFields(
			fields.New("yes", fields.TypeBool,
				fields.WithHelp("Skip confirmation prompt"),
				fields.WithDefault(false),
				fields.WithShortFlag("y"),
			),
			fields.New("force", fields.TypeBool,
				fields.WithHelp("Apply even if dependencies are missing"),
				fields.WithDefault(false),
				fields.WithShortFlag("f"),
			),
		),
	)
	if err != nil {
		return nil, err
	}

	s := schema.NewSchema(schema.WithSections(defaultSection))

	cmdDef := cmds.NewCommandDefinition(
		"apply",
		cmds.WithShort("Apply a tactic to create new nodes"),
		cmds.WithSchema(s),
	)

	return &ApplyCommand{CommandDefinition: cmdDef}, nil
}

var _ cmds.BareCommand = &ApplyCommand{}

func (c *ApplyCommand) Run(ctx context.Context, vals *values.Values) error {
	_ = ctx
	_ = vals
	return errors.New("apply: not implemented yet")
}
