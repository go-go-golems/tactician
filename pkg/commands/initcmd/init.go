package initcmd

import (
	"context"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/go-go-golems/tactician/pkg/commands/sections"
	"github.com/go-go-golems/tactician/pkg/db"
	"github.com/go-go-golems/tactician/pkg/defaults"
	"github.com/go-go-golems/tactician/pkg/store"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
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
	settings := &sections.TacticianSettings{}
	if err := values.DecodeSectionInto(vals, sections.TacticianSlug, settings); err != nil {
		return errors.Wrap(err, "decode tactician settings")
	}

	if err := store.InitDir(settings.Dir); err != nil {
		return err
	}

	var tactics []*db.Tactic
	if err := yaml.Unmarshal(defaults.DefaultTacticsYAML, &tactics); err != nil {
		return errors.Wrap(err, "parse embedded default tactics")
	}
	if err := store.SeedTacticsIfMissing(settings.Dir, tactics); err != nil {
		return err
	}

	// TODO(manuel): Optionally set project metadata (name/root_goal).
	return nil
}
