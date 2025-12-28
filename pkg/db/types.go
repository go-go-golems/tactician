package db

import (
	"encoding/json"
	"time"
)

type Node struct {
	ID           string          `json:"id"`
	Type         string          `json:"type"`
	Output       string          `json:"output"`
	Status       string          `json:"status"`
	CreatedBy    *string         `json:"created_by,omitempty"`
	CreatedAt    time.Time       `json:"created_at"`
	CompletedAt  *time.Time      `json:"completed_at,omitempty"`
	ParentTactic *string         `json:"parent_tactic,omitempty"`
	IntroducedAs *string         `json:"introduced_as,omitempty"`
	Data         json.RawMessage `json:"data,omitempty"`
}

type Edge struct {
	SourceNodeID string `json:"source_node_id"`
	TargetNodeID string `json:"target_node_id"`
}

type ActionLogEntry struct {
	ID        int64     `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	Action    string    `json:"action"`
	Details   *string   `json:"details,omitempty"`
	NodeID    *string   `json:"node_id,omitempty"`
	TacticID  *string   `json:"tactic_id,omitempty"`
}

type SessionSummary struct {
	TotalActions   int            `json:"total_actions"`
	NodesCreated   int            `json:"nodes_created"`
	NodesCompleted int            `json:"nodes_completed"`
	TacticsApplied int            `json:"tactics_applied"`
	NodesModified  int            `json:"nodes_modified"`
	ActionsByType  map[string]int `json:"actions_by_type"`
}
