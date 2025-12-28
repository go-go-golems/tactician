package db

import (
	"context"
	"testing"
)

func TestTacticsDB_AddTactic_SubtasksFromData(t *testing.T) {
	ctx := context.Background()

	sqlDB, err := OpenSQLiteMemory(ctx)
	if err != nil {
		t.Fatalf("OpenSQLiteMemory: %v", err)
	}
	defer func() { _ = sqlDB.Close() }()

	tdb := NewTacticsDBFromDB(sqlDB)
	if err := tdb.InitSchema(ctx); err != nil {
		t.Fatalf("InitSchema: %v", err)
	}

	// Subtasks provided via data.subtasks (JS compatibility path).
	tactic := &Tactic{
		ID:     "t1",
		Type:   "document",
		Output: "out",
		Data: map[string]interface{}{
			"subtasks": []interface{}{
				map[string]interface{}{
					"id":         "s1",
					"type":       "document",
					"output":     "o1",
					"depends_on": []interface{}{"s0"},
				},
			},
		},
	}

	if err := tdb.AddTactic(ctx, tactic); err != nil {
		t.Fatalf("AddTactic: %v", err)
	}

	got, err := tdb.GetTactic(ctx, "t1")
	if err != nil {
		t.Fatalf("GetTactic: %v", err)
	}
	if got == nil {
		t.Fatalf("expected tactic")
	}
	if len(got.Subtasks) != 1 {
		t.Fatalf("expected 1 subtask, got %d", len(got.Subtasks))
	}
	if got.Subtasks[0].ID != "s1" {
		t.Fatalf("expected subtask id s1, got %q", got.Subtasks[0].ID)
	}
	if len(got.Subtasks[0].DependsOn) != 1 || got.Subtasks[0].DependsOn[0] != "s0" {
		t.Fatalf("expected depends_on [s0], got %#v", got.Subtasks[0].DependsOn)
	}
}


