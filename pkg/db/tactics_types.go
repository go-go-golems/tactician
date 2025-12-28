package db

// Disk + API model for tactics.
// This matches the JavaScript YAML shape (one file per tactic) and the JS DB schema.

type Tactic struct {
	ID          string                 `yaml:"id"`
	Type        string                 `yaml:"type"`
	Output      string                 `yaml:"output"`
	Description string                 `yaml:"description,omitempty"`
	Tags        []string               `yaml:"tags,omitempty"`
	Match       []string               `yaml:"match,omitempty"`
	Premises    []string               `yaml:"premises,omitempty"`
	Subtasks    []TacticSubtask        `yaml:"subtasks,omitempty"`
	Data        map[string]interface{} `yaml:"data,omitempty"`
}

type TacticSubtask struct {
	ID        string                 `yaml:"id"`
	Output    string                 `yaml:"output"`
	Type      string                 `yaml:"type"`
	DependsOn []string               `yaml:"depends_on,omitempty"`
	Data      map[string]interface{} `yaml:"data,omitempty"`
}
