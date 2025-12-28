package sections

import (
	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
)

const ProjectSlug = "project"

type ProjectSettings struct {
	ProjectDBPath string `glazed.parameter:"project-db-path"`
	TacticsDBPath string `glazed.parameter:"tactics-db-path"`
}

func NewProjectSection() (*schema.SectionImpl, error) {
	return schema.NewSection(
		ProjectSlug,
		"Project",
		schema.WithDescription("Project database configuration"),
		schema.WithFields(
			fields.New("project-db-path", fields.TypeString,
				fields.WithHelp("Path to the project database (nodes, edges, action log, meta)"),
				fields.WithDefault(".tactician/project.db"),
			),
			fields.New("tactics-db-path", fields.TypeString,
				fields.WithHelp("Path to the tactics database (tactics, dependencies, subtasks)"),
				fields.WithDefault(".tactician/tactics.db"),
			),
		),
	)
}
