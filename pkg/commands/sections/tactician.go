package sections

import (
	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
)

const TacticianSlug = "tactician"

type TacticianSettings struct {
	Dir string `glazed.parameter:"dir"`
}

func NewTacticianSection() (*schema.SectionImpl, error) {
	return schema.NewSection(
		TacticianSlug,
		"Tactician",
		schema.WithPrefix("tactician-"),
		schema.WithDescription("Tactician storage settings (.tactician/ YAML source-of-truth)"),
		schema.WithFields(
			fields.New("dir", fields.TypeString,
				fields.WithHelp("Path to the tactician directory (contains YAML state)"),
				fields.WithDefault(".tactician"),
			),
		),
	)
}
