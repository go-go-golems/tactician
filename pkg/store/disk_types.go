package store

import (
	"time"
)

// Disk layout (YAML source-of-truth).
//
// NOTE: This is intentionally separate from the sqlite schema; it is the persisted representation.

type diskProjectFile struct {
	Project diskProjectMeta `yaml:"project"`
	Nodes   []diskNode      `yaml:"nodes"`
	Edges   []diskEdge      `yaml:"edges"`
}

type diskProjectMeta struct {
	Name     string `yaml:"name"`
	RootGoal string `yaml:"root_goal"`
}

type diskNode struct {
	ID           string                 `yaml:"id"`
	Type         string                 `yaml:"type"`
	Output       string                 `yaml:"output"`
	Status       string                 `yaml:"status"`
	CreatedBy    *string                `yaml:"created_by,omitempty"`
	CreatedAt    *time.Time             `yaml:"created_at,omitempty"`
	CompletedAt  *time.Time             `yaml:"completed_at,omitempty"`
	ParentTactic *string                `yaml:"parent_tactic,omitempty"`
	IntroducedAs *string                `yaml:"introduced_as,omitempty"`
	Data         map[string]interface{} `yaml:"data,omitempty"`
}

type diskEdge struct {
	Source string `yaml:"source"`
	Target string `yaml:"target"`
}

type diskActionLogFile []diskActionLogEntry

type diskActionLogEntry struct {
	Timestamp time.Time `yaml:"timestamp"`
	Action    string    `yaml:"action"`
	Details   *string   `yaml:"details,omitempty"`
	NodeID    *string   `yaml:"node_id,omitempty"`
	TacticID  *string   `yaml:"tactic_id,omitempty"`
}
