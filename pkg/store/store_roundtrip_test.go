package store

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/go-go-golems/tactician/pkg/db"
)

func TestRoundTrip_ProjectActionLogAndTactics(t *testing.T) {
	ctx := context.Background()
	base := t.TempDir()
	dir := filepath.Join(base, ".tactician")

	if err := InitDir(dir); err != nil {
		t.Fatalf("InitDir: %v", err)
	}

	// Seed a tactic file (one-file-per-tactic).
	if err := SeedTacticsIfMissing(dir, []*db.Tactic{
		{
			ID:          "t1",
			Type:        "document",
			Output:      "doc-1",
			Description: "test tactic",
			Tags:        []string{"test"},
			Match:       []string{},
		},
	}); err != nil {
		t.Fatalf("SeedTacticsIfMissing: %v", err)
	}

	st, err := Load(ctx, dir)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	// Validate tactic import
	tactic, err := st.Tactics.GetTactic(ctx, "t1")
	if err != nil {
		_ = st.Close()
		t.Fatalf("GetTactic: %v", err)
	}
	if tactic == nil || tactic.ID != "t1" {
		_ = st.Close()
		t.Fatalf("expected tactic t1 to be loaded")
	}

	// Mutate project state (node + action log), then save.
	if err := st.Project.AddNode(ctx, &db.Node{
		ID:     "root",
		Type:   "project_artifact",
		Output: "README.md",
		Status: "pending",
	}); err != nil {
		_ = st.Close()
		t.Fatalf("AddNode: %v", err)
	}

	details := "Created node: root"
	nodeID := "root"
	if err := st.Project.LogAction(ctx, "node_created", &details, &nodeID, nil); err != nil {
		_ = st.Close()
		t.Fatalf("LogAction: %v", err)
	}

	st.Dirty = true
	if err := st.Save(ctx); err != nil {
		_ = st.Close()
		t.Fatalf("Save: %v", err)
	}
	_ = st.Close()

	// Reload and assert state persisted to YAML.
	st2, err := Load(ctx, dir)
	if err != nil {
		t.Fatalf("Load (2): %v", err)
	}
	defer func() { _ = st2.Close() }()

	n, err := st2.Project.GetNode(ctx, "root")
	if err != nil {
		t.Fatalf("GetNode: %v", err)
	}
	if n == nil || n.ID != "root" {
		t.Fatalf("expected node root after reload")
	}

	logs, err := st2.Project.GetActionLog(ctx, nil, nil)
	if err != nil {
		t.Fatalf("GetActionLog: %v", err)
	}
	if len(logs) == 0 {
		t.Fatalf("expected action log entries after reload")
	}
}


