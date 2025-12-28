package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/go-go-golems/tactician/pkg/db"
	"github.com/pkg/errors"
)

type State struct {
	Dir string

	SQL     *sql.DB
	Project *db.ProjectDB
	Tactics *db.TacticsDB

	// TODO(manuel): Add TacticsDB once implemented.
	Dirty bool
}

func Load(ctx context.Context, tacticianDir string) (*State, error) {
	// Treat tacticianDir as a filesystem directory path.
	// Commands should pass settings.Dir (default: ".tactician").
	if tacticianDir == "" {
		return nil, errors.New("tactician dir is empty")
	}

	if _, err := os.Stat(tacticianDir); err != nil {
		return nil, errors.Wrap(err, "stat tactician dir")
	}
	if _, err := os.Stat(filepath.Join(tacticianDir, projectFileName)); err != nil {
		return nil, errors.Wrap(err, "stat project.yaml (project not initialized?)")
	}

	sqlDB, err := db.OpenSQLiteMemory(ctx)
	if err != nil {
		return nil, err
	}

	projectDB := db.NewProjectDBFromDB(sqlDB)
	if err := projectDB.InitSchema(ctx); err != nil {
		_ = sqlDB.Close()
		return nil, err
	}

	tacticsDB := db.NewTacticsDBFromDB(sqlDB)
	if err := tacticsDB.InitSchema(ctx); err != nil {
		_ = sqlDB.Close()
		return nil, err
	}

	s := &State{
		Dir:     tacticianDir,
		SQL:     sqlDB,
		Project: projectDB,
		Tactics: tacticsDB,
		Dirty:   false,
	}

	if err := s.importFromDisk(ctx); err != nil {
		_ = s.Close()
		return nil, err
	}

	return s, nil
}

func (s *State) Close() error {
	if s == nil || s.SQL == nil {
		return nil
	}
	err := s.SQL.Close()
	s.SQL = nil
	return err
}

func (s *State) Save(ctx context.Context) error {
	if s == nil {
		return errors.New("nil state")
	}
	if !s.Dirty {
		return nil
	}

	if err := s.exportToDisk(ctx); err != nil {
		return err
	}
	return nil
}

func (s *State) importFromDisk(ctx context.Context) error {
	// Project
	project, err := readProjectFile(s.Dir)
	if err != nil {
		return err
	}

	if err := s.Project.SetProjectMeta(ctx, project.Project.Name, project.Project.RootGoal); err != nil {
		return err
	}

	for _, n := range project.Nodes {
		var data json.RawMessage
		if len(n.Data) > 0 {
			b, err := json.Marshal(n.Data)
			if err != nil {
				return errors.Wrap(err, "marshal node data")
			}
			data = b
		}

		node := &db.Node{
			ID:           n.ID,
			Type:         n.Type,
			Output:       n.Output,
			Status:       n.Status,
			CreatedBy:    n.CreatedBy,
			ParentTactic: n.ParentTactic,
			IntroducedAs: n.IntroducedAs,
			Data:         data,
		}
		if n.CreatedAt != nil {
			node.CreatedAt = n.CreatedAt.UTC()
		}
		if n.CompletedAt != nil {
			t := n.CompletedAt.UTC()
			node.CompletedAt = &t
		}

		if err := s.Project.AddNode(ctx, node); err != nil {
			return err
		}
	}

	for _, e := range project.Edges {
		if err := s.Project.AddEdge(ctx, e.Source, e.Target); err != nil {
			return err
		}
	}

	// Action log
	logEntries, err := readActionLogFile(s.Dir)
	if err != nil {
		return err
	}
	for _, e := range logEntries {
		// Preserve timestamp from disk.
		_, err := s.SQL.ExecContext(ctx, `
INSERT INTO action_log (timestamp, action, details, node_id, tactic_id)
VALUES (?, ?, ?, ?, ?)
`, e.Timestamp.UTC().Format(time.RFC3339Nano), e.Action, e.Details, e.NodeID, e.TacticID)
		if err != nil {
			return errors.Wrap(err, "import action log")
		}
	}

	// Tactics (one file per tactic)
	if _, err := os.Stat(tacticsDirPath(s.Dir)); err == nil {
		tactics, err := readTacticsDir(s.Dir)
		if err != nil {
			return err
		}
		for _, t := range tactics {
			if err := s.Tactics.AddTactic(ctx, t); err != nil {
				return err
			}
		}
	} else if !os.IsNotExist(err) {
		return errors.Wrap(err, "stat tactics dir")
	}

	return nil
}

func (s *State) exportToDisk(ctx context.Context) error {
	meta, err := s.Project.GetProjectMeta(ctx)
	if err != nil {
		return err
	}
	nodes, err := s.Project.GetAllNodes(ctx)
	if err != nil {
		return err
	}
	edges, err := s.Project.GetEdges(ctx)
	if err != nil {
		return err
	}

	project := &diskProjectFile{
		Project: diskProjectMeta{
			Name:     meta["name"],
			RootGoal: meta["root_goal"],
		},
		Nodes: []diskNode{},
		Edges: []diskEdge{},
	}

	for _, n := range nodes {
		var data map[string]interface{}
		if len(n.Data) > 0 {
			if err := json.Unmarshal(n.Data, &data); err != nil {
				return errors.Wrap(err, "unmarshal node data")
			}
		}

		createdAt := n.CreatedAt
		node := diskNode{
			ID:           n.ID,
			Type:         n.Type,
			Output:       n.Output,
			Status:       n.Status,
			CreatedBy:    n.CreatedBy,
			ParentTactic: n.ParentTactic,
			IntroducedAs: n.IntroducedAs,
			Data:         data,
		}
		if !createdAt.IsZero() {
			t := createdAt.UTC()
			node.CreatedAt = &t
		}
		if n.CompletedAt != nil && !n.CompletedAt.IsZero() {
			t := n.CompletedAt.UTC()
			node.CompletedAt = &t
		}

		project.Nodes = append(project.Nodes, node)
	}

	for _, e := range edges {
		project.Edges = append(project.Edges, diskEdge{Source: e.SourceNodeID, Target: e.TargetNodeID})
	}

	if err := writeProjectFile(s.Dir, project); err != nil {
		return err
	}

	logs, err := s.Project.GetActionLog(ctx, nil, nil)
	if err != nil {
		return err
	}
	outLog := diskActionLogFile{}
	for _, l := range logs {
		outLog = append(outLog, diskActionLogEntry{
			Timestamp: l.Timestamp.UTC(),
			Action:    l.Action,
			Details:   l.Details,
			NodeID:    l.NodeID,
			TacticID:  l.TacticID,
		})
	}
	if err := writeActionLogFile(s.Dir, outLog); err != nil {
		return err
	}

	tactics, err := s.Tactics.GetAllTactics(ctx)
	if err != nil {
		return err
	}
	if err := writeTacticsDir(s.Dir, tactics); err != nil {
		return err
	}

	return nil
}
