package initcmd

import (
	"context"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/go-go-golems/tactician/pkg/commands/sections"
	"github.com/pkg/errors"
)

type InitCommand struct {
	*cmds.CommandDefinition
}

func NewInitCommand() (*InitCommand, error) {
	tacticianSection, err := sections.NewTacticianSection()
	if err != nil {
		return nil, err
	}

	s := schema.NewSchema(
		schema.WithSections(tacticianSection),
	)

	cmdDef := cmds.NewCommandDefinition(
		"init",
		cmds.WithShort("Initialize a new Tactician project"),
		cmds.WithLong("Creates .tactician/ directory and initializes project and tactics databases"),
		cmds.WithSchema(s),
	)

	return &InitCommand{CommandDefinition: cmdDef}, nil
}

var _ cmds.BareCommand = &InitCommand{}

func (c *InitCommand) Run(ctx context.Context, vals *values.Values) error {
	_ = ctx
	_ = vals
	return errors.New("init: not implemented yet")
}
