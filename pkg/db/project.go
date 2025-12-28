package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

type ProjectDB struct {
	DBPath string
	db     *sql.DB
}

func NewProjectDB(dbPath string) *ProjectDB {
	return &ProjectDB{DBPath: dbPath}
}

func NewProjectDBFromDB(db *sql.DB) *ProjectDB {
	return &ProjectDB{db: db}
}

func (p *ProjectDB) Open(ctx context.Context) error {
	if p.db != nil {
		return nil
	}
	db, err := openSQLite(ctx, p.DBPath)
	if err != nil {
		return err
	}
	p.db = db
	return nil
}

func (p *ProjectDB) Close() error {
	if p.db == nil {
		return nil
	}
	err := p.db.Close()
	p.db = nil
	return err
}

func (p *ProjectDB) InitSchema(ctx context.Context) error {
	if p.db == nil {
		return errors.New("project db not open")
	}

	const schemaSQL = `
CREATE TABLE IF NOT EXISTS project (
  key TEXT PRIMARY KEY,
  value TEXT
);

CREATE TABLE IF NOT EXISTS nodes (
  id TEXT PRIMARY KEY,
  type TEXT NOT NULL,
  output TEXT NOT NULL,
  status TEXT NOT NULL DEFAULT 'pending',
  created_by TEXT,
  created_at TEXT NOT NULL,
  completed_at TEXT,
  parent_tactic TEXT,
  introduced_as TEXT,
  data TEXT
);

CREATE TABLE IF NOT EXISTS edges (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  source_node_id TEXT NOT NULL,
  target_node_id TEXT NOT NULL,
  FOREIGN KEY (source_node_id) REFERENCES nodes(id) ON DELETE CASCADE,
  FOREIGN KEY (target_node_id) REFERENCES nodes(id) ON DELETE CASCADE,
  UNIQUE(source_node_id, target_node_id)
);

CREATE INDEX IF NOT EXISTS idx_edges_source ON edges(source_node_id);
CREATE INDEX IF NOT EXISTS idx_edges_target ON edges(target_node_id);
CREATE INDEX IF NOT EXISTS idx_nodes_status ON nodes(status);

CREATE TABLE IF NOT EXISTS action_log (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  timestamp TEXT NOT NULL,
  action TEXT NOT NULL,
  details TEXT,
  node_id TEXT,
  tactic_id TEXT
);

CREATE INDEX IF NOT EXISTS idx_log_timestamp ON action_log(timestamp);
CREATE INDEX IF NOT EXISTS idx_log_action ON action_log(action);
`

	_, err := p.db.ExecContext(ctx, schemaSQL)
	return errors.Wrap(err, "init project schema")
}

func (p *ProjectDB) SetProjectMeta(ctx context.Context, name string, rootGoal string) error {
	if p.db == nil {
		return errors.New("project db not open")
	}

	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return errors.Wrap(err, "begin tx")
	}
	defer func() { _ = tx.Rollback() }()

	stmt, err := tx.PrepareContext(ctx, "INSERT OR REPLACE INTO project (key, value) VALUES (?, ?)")
	if err != nil {
		return errors.Wrap(err, "prepare project meta upsert")
	}
	defer func() { _ = stmt.Close() }()

	if _, err := stmt.ExecContext(ctx, "name", name); err != nil {
		return errors.Wrap(err, "set project name")
	}
	if _, err := stmt.ExecContext(ctx, "root_goal", rootGoal); err != nil {
		return errors.Wrap(err, "set project root_goal")
	}

	if err := tx.Commit(); err != nil {
		return errors.Wrap(err, "commit tx")
	}
	return nil
}

func (p *ProjectDB) GetProjectMeta(ctx context.Context) (map[string]string, error) {
	if p.db == nil {
		return nil, errors.New("project db not open")
	}

	rows, err := p.db.QueryContext(ctx, "SELECT key, value FROM project")
	if err != nil {
		return nil, errors.Wrap(err, "select project meta")
	}
	defer func() { _ = rows.Close() }()

	meta := map[string]string{}
	for rows.Next() {
		var k, v string
		if err := rows.Scan(&k, &v); err != nil {
			return nil, errors.Wrap(err, "scan project meta")
		}
		meta[k] = v
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "iterate project meta")
	}
	return meta, nil
}

func (p *ProjectDB) AddNode(ctx context.Context, node *Node) error {
	if p.db == nil {
		return errors.New("project db not open")
	}
	if node == nil {
		return errors.New("node is nil")
	}
	if node.CreatedAt.IsZero() {
		node.CreatedAt = time.Now()
	}
	status := node.Status
	if status == "" {
		status = "pending"
	}

	var completedAt *string
	if node.CompletedAt != nil && !node.CompletedAt.IsZero() {
		s := node.CompletedAt.UTC().Format(time.RFC3339Nano)
		completedAt = &s
	}

	var data *string
	if len(node.Data) > 0 {
		s := string(node.Data)
		data = &s
	}

	_, err := p.db.ExecContext(ctx, `
INSERT INTO nodes (id, type, output, status, created_by, created_at, completed_at, parent_tactic, introduced_as, data)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
`,
		node.ID,
		node.Type,
		node.Output,
		status,
		node.CreatedBy,
		node.CreatedAt.UTC().Format(time.RFC3339Nano),
		completedAt,
		node.ParentTactic,
		node.IntroducedAs,
		data,
	)
	return errors.Wrap(err, "insert node")
}

func (p *ProjectDB) GetNode(ctx context.Context, id string) (*Node, error) {
	if p.db == nil {
		return nil, errors.New("project db not open")
	}

	row := p.db.QueryRowContext(ctx, `
SELECT id, type, output, status, created_by, created_at, completed_at, parent_tactic, introduced_as, data
FROM nodes WHERE id = ?
`, id)

	var n Node
	var createdBy, createdAt, completedAt, parentTactic, introducedAs, data sql.NullString
	if err := row.Scan(
		&n.ID, &n.Type, &n.Output, &n.Status,
		&createdBy, &createdAt, &completedAt, &parentTactic, &introducedAs, &data,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, errors.Wrap(err, "scan node")
	}

	if createdBy.Valid {
		n.CreatedBy = &createdBy.String
	}
	if createdAt.Valid {
		t, err := time.Parse(time.RFC3339Nano, createdAt.String)
		if err != nil {
			return nil, errors.Wrap(err, "parse created_at")
		}
		n.CreatedAt = t
	}
	if completedAt.Valid {
		t, err := time.Parse(time.RFC3339Nano, completedAt.String)
		if err != nil {
			return nil, errors.Wrap(err, "parse completed_at")
		}
		n.CompletedAt = &t
	}
	if parentTactic.Valid {
		n.ParentTactic = &parentTactic.String
	}
	if introducedAs.Valid {
		n.IntroducedAs = &introducedAs.String
	}
	if data.Valid && data.String != "" {
		n.Data = json.RawMessage(data.String)
	}

	return &n, nil
}

func (p *ProjectDB) GetAllNodes(ctx context.Context) ([]*Node, error) {
	if p.db == nil {
		return nil, errors.New("project db not open")
	}

	rows, err := p.db.QueryContext(ctx, `
SELECT id, type, output, status, created_by, created_at, completed_at, parent_tactic, introduced_as, data
FROM nodes
`)
	if err != nil {
		return nil, errors.Wrap(err, "select nodes")
	}
	defer func() { _ = rows.Close() }()

	var ret []*Node
	for rows.Next() {
		var n Node
		var createdBy, createdAt, completedAt, parentTactic, introducedAs, data sql.NullString
		if err := rows.Scan(
			&n.ID, &n.Type, &n.Output, &n.Status,
			&createdBy, &createdAt, &completedAt, &parentTactic, &introducedAs, &data,
		); err != nil {
			return nil, errors.Wrap(err, "scan node")
		}
		if createdBy.Valid {
			n.CreatedBy = &createdBy.String
		}
		if createdAt.Valid && createdAt.String != "" {
			t, err := time.Parse(time.RFC3339Nano, createdAt.String)
			if err != nil {
				return nil, errors.Wrap(err, "parse created_at")
			}
			n.CreatedAt = t
		}
		if completedAt.Valid && completedAt.String != "" {
			t, err := time.Parse(time.RFC3339Nano, completedAt.String)
			if err != nil {
				return nil, errors.Wrap(err, "parse completed_at")
			}
			n.CompletedAt = &t
		}
		if parentTactic.Valid {
			n.ParentTactic = &parentTactic.String
		}
		if introducedAs.Valid {
			n.IntroducedAs = &introducedAs.String
		}
		if data.Valid && data.String != "" {
			n.Data = json.RawMessage(data.String)
		}

		ret = append(ret, &n)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "iterate nodes")
	}
	return ret, nil
}

func (p *ProjectDB) UpdateNodeStatus(ctx context.Context, id string, status string, completedAt *time.Time) error {
	if p.db == nil {
		return errors.New("project db not open")
	}

	var completedAtStr any
	if completedAt != nil && !completedAt.IsZero() {
		completedAtStr = completedAt.UTC().Format(time.RFC3339Nano)
	} else {
		completedAtStr = nil
	}

	_, err := p.db.ExecContext(ctx, "UPDATE nodes SET status = ?, completed_at = ? WHERE id = ?", status, completedAtStr, id)
	return errors.Wrap(err, "update node status")
}

func (p *ProjectDB) DeleteNode(ctx context.Context, id string) error {
	if p.db == nil {
		return errors.New("project db not open")
	}
	_, err := p.db.ExecContext(ctx, "DELETE FROM nodes WHERE id = ?", id)
	return errors.Wrap(err, "delete node")
}

func (p *ProjectDB) AddEdge(ctx context.Context, sourceID, targetID string) error {
	if p.db == nil {
		return errors.New("project db not open")
	}
	_, err := p.db.ExecContext(ctx, "INSERT OR IGNORE INTO edges (source_node_id, target_node_id) VALUES (?, ?)", sourceID, targetID)
	return errors.Wrap(err, "insert edge")
}

func (p *ProjectDB) GetEdges(ctx context.Context) ([]Edge, error) {
	if p.db == nil {
		return nil, errors.New("project db not open")
	}

	rows, err := p.db.QueryContext(ctx, "SELECT source_node_id, target_node_id FROM edges")
	if err != nil {
		return nil, errors.Wrap(err, "select edges")
	}
	defer func() { _ = rows.Close() }()

	var edges []Edge
	for rows.Next() {
		var e Edge
		if err := rows.Scan(&e.SourceNodeID, &e.TargetNodeID); err != nil {
			return nil, errors.Wrap(err, "scan edge")
		}
		edges = append(edges, e)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "iterate edges")
	}
	return edges, nil
}

func (p *ProjectDB) GetDependencies(ctx context.Context, nodeID string) ([]*Node, error) {
	if p.db == nil {
		return nil, errors.New("project db not open")
	}

	rows, err := p.db.QueryContext(ctx, `
SELECT n.id, n.type, n.output, n.status, n.created_by, n.created_at, n.completed_at, n.parent_tactic, n.introduced_as, n.data
FROM nodes n
INNER JOIN edges e ON e.source_node_id = n.id
WHERE e.target_node_id = ?
`, nodeID)
	if err != nil {
		return nil, errors.Wrap(err, "select dependencies")
	}
	defer func() { _ = rows.Close() }()

	var ret []*Node
	for rows.Next() {
		var n Node
		var createdBy, createdAt, completedAt, parentTactic, introducedAs, data sql.NullString
		if err := rows.Scan(
			&n.ID, &n.Type, &n.Output, &n.Status,
			&createdBy, &createdAt, &completedAt, &parentTactic, &introducedAs, &data,
		); err != nil {
			return nil, errors.Wrap(err, "scan dependency node")
		}
		if createdBy.Valid {
			n.CreatedBy = &createdBy.String
		}
		if createdAt.Valid && createdAt.String != "" {
			t, err := time.Parse(time.RFC3339Nano, createdAt.String)
			if err != nil {
				return nil, errors.Wrap(err, "parse created_at")
			}
			n.CreatedAt = t
		}
		if completedAt.Valid && completedAt.String != "" {
			t, err := time.Parse(time.RFC3339Nano, completedAt.String)
			if err != nil {
				return nil, errors.Wrap(err, "parse completed_at")
			}
			n.CompletedAt = &t
		}
		if parentTactic.Valid {
			n.ParentTactic = &parentTactic.String
		}
		if introducedAs.Valid {
			n.IntroducedAs = &introducedAs.String
		}
		if data.Valid && data.String != "" {
			n.Data = json.RawMessage(data.String)
		}
		ret = append(ret, &n)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "iterate dependencies")
	}
	return ret, nil
}

func (p *ProjectDB) GetBlockedBy(ctx context.Context, nodeID string) ([]*Node, error) {
	if p.db == nil {
		return nil, errors.New("project db not open")
	}

	rows, err := p.db.QueryContext(ctx, `
SELECT n.id, n.type, n.output, n.status, n.created_by, n.created_at, n.completed_at, n.parent_tactic, n.introduced_as, n.data
FROM nodes n
INNER JOIN edges e ON e.target_node_id = n.id
WHERE e.source_node_id = ?
`, nodeID)
	if err != nil {
		return nil, errors.Wrap(err, "select blocked-by")
	}
	defer func() { _ = rows.Close() }()

	var ret []*Node
	for rows.Next() {
		var n Node
		var createdBy, createdAt, completedAt, parentTactic, introducedAs, data sql.NullString
		if err := rows.Scan(
			&n.ID, &n.Type, &n.Output, &n.Status,
			&createdBy, &createdAt, &completedAt, &parentTactic, &introducedAs, &data,
		); err != nil {
			return nil, errors.Wrap(err, "scan blocked-by node")
		}
		if createdBy.Valid {
			n.CreatedBy = &createdBy.String
		}
		if createdAt.Valid && createdAt.String != "" {
			t, err := time.Parse(time.RFC3339Nano, createdAt.String)
			if err != nil {
				return nil, errors.Wrap(err, "parse created_at")
			}
			n.CreatedAt = t
		}
		if completedAt.Valid && completedAt.String != "" {
			t, err := time.Parse(time.RFC3339Nano, completedAt.String)
			if err != nil {
				return nil, errors.Wrap(err, "parse completed_at")
			}
			n.CompletedAt = &t
		}
		if parentTactic.Valid {
			n.ParentTactic = &parentTactic.String
		}
		if introducedAs.Valid {
			n.IntroducedAs = &introducedAs.String
		}
		if data.Valid && data.String != "" {
			n.Data = json.RawMessage(data.String)
		}
		ret = append(ret, &n)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "iterate blocked-by")
	}
	return ret, nil
}

func (p *ProjectDB) LogAction(ctx context.Context, action string, details *string, nodeID *string, tacticID *string) error {
	if p.db == nil {
		return errors.New("project db not open")
	}

	_, err := p.db.ExecContext(ctx, `
INSERT INTO action_log (timestamp, action, details, node_id, tactic_id)
VALUES (?, ?, ?, ?, ?)
`, time.Now().UTC().Format(time.RFC3339Nano), action, details, nodeID, tacticID)
	return errors.Wrap(err, "insert action_log")
}

func (p *ProjectDB) GetActionLog(ctx context.Context, limit *int, since *time.Time) ([]ActionLogEntry, error) {
	if p.db == nil {
		return nil, errors.New("project db not open")
	}

	query := "SELECT id, timestamp, action, details, node_id, tactic_id FROM action_log"
	var args []any
	if since != nil && !since.IsZero() {
		query += " WHERE timestamp >= ?"
		args = append(args, since.UTC().Format(time.RFC3339Nano))
	}
	query += " ORDER BY timestamp DESC"
	if limit != nil && *limit > 0 {
		query += " LIMIT ?"
		args = append(args, *limit)
	}

	rows, err := p.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, errors.Wrap(err, "select action_log")
	}
	defer func() { _ = rows.Close() }()

	var ret []ActionLogEntry
	for rows.Next() {
		var e ActionLogEntry
		var ts sql.NullString
		var details, nodeID, tacticID sql.NullString
		if err := rows.Scan(&e.ID, &ts, &e.Action, &details, &nodeID, &tacticID); err != nil {
			return nil, errors.Wrap(err, "scan action_log")
		}
		if ts.Valid && ts.String != "" {
			t, err := time.Parse(time.RFC3339Nano, ts.String)
			if err != nil {
				return nil, errors.Wrap(err, "parse log timestamp")
			}
			e.Timestamp = t
		}
		if details.Valid {
			e.Details = &details.String
		}
		if nodeID.Valid {
			e.NodeID = &nodeID.String
		}
		if tacticID.Valid {
			e.TacticID = &tacticID.String
		}
		ret = append(ret, e)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "iterate action_log")
	}

	return ret, nil
}

func (p *ProjectDB) GetSessionSummary(ctx context.Context, since *time.Time) (*SessionSummary, error) {
	logs, err := p.GetActionLog(ctx, nil, since)
	if err != nil {
		return nil, err
	}

	summary := &SessionSummary{
		TotalActions:  len(logs),
		ActionsByType: map[string]int{},
	}

	for _, l := range logs {
		summary.ActionsByType[l.Action] = summary.ActionsByType[l.Action] + 1
		switch l.Action {
		case "node_created":
			summary.NodesCreated++
		case "node_completed":
			summary.NodesCompleted++
		case "tactic_applied":
			summary.TacticsApplied++
		case "node_updated":
			summary.NodesModified++
		}
	}

	return summary, nil
}

type projectYAML struct {
	Project struct {
		Name     string `yaml:"name"`
		RootGoal string `yaml:"root_goal"`
	} `yaml:"project"`
	Nodes map[string]nodeYAML `yaml:"nodes"`
}

type nodeYAML struct {
	Type         string                 `yaml:"type"`
	Output       string                 `yaml:"output"`
	Status       string                 `yaml:"status"`
	CreatedBy    *string                `yaml:"created_by,omitempty"`
	CreatedAt    *string                `yaml:"created_at,omitempty"`
	CompletedAt  *string                `yaml:"completed_at,omitempty"`
	ParentTactic *string                `yaml:"parent_tactic,omitempty"`
	IntroducedAs *string                `yaml:"introduced_as,omitempty"`
	Dependencies *nodeDepsYAML          `yaml:"dependencies,omitempty"`
	Blocks       []string               `yaml:"blocks,omitempty"`
	Data         map[string]interface{} `yaml:"data,omitempty"`
}

type nodeDepsYAML struct {
	Match []string `yaml:"match,omitempty"`
}

func (p *ProjectDB) ExportToYAML(ctx context.Context) (string, error) {
	meta, err := p.GetProjectMeta(ctx)
	if err != nil {
		return "", err
	}
	nodes, err := p.GetAllNodes(ctx)
	if err != nil {
		return "", err
	}

	out := projectYAML{
		Nodes: map[string]nodeYAML{},
	}
	out.Project.Name = meta["name"]
	out.Project.RootGoal = meta["root_goal"]

	for _, n := range nodes {
		entry := nodeYAML{
			Type:         n.Type,
			Output:       n.Output,
			Status:       n.Status,
			CreatedBy:    n.CreatedBy,
			ParentTactic: n.ParentTactic,
			IntroducedAs: n.IntroducedAs,
		}
		if !n.CreatedAt.IsZero() {
			s := n.CreatedAt.UTC().Format(time.RFC3339Nano)
			entry.CreatedAt = &s
		}
		if n.CompletedAt != nil && !n.CompletedAt.IsZero() {
			s := n.CompletedAt.UTC().Format(time.RFC3339Nano)
			entry.CompletedAt = &s
		}

		deps, err := p.GetDependencies(ctx, n.ID)
		if err != nil {
			return "", err
		}
		if len(deps) > 0 {
			entry.Dependencies = &nodeDepsYAML{Match: make([]string, 0, len(deps))}
			for _, d := range deps {
				entry.Dependencies.Match = append(entry.Dependencies.Match, d.ID)
			}
		}

		blocks, err := p.GetBlockedBy(ctx, n.ID)
		if err != nil {
			return "", err
		}
		if len(blocks) > 0 {
			entry.Blocks = make([]string, 0, len(blocks))
			for _, b := range blocks {
				entry.Blocks = append(entry.Blocks, b.ID)
			}
		}

		if len(n.Data) > 0 {
			var m map[string]interface{}
			if err := json.Unmarshal(n.Data, &m); err != nil {
				return "", errors.Wrap(err, "unmarshal node data for yaml export")
			}
			entry.Data = m
		}

		out.Nodes[n.ID] = entry
	}

	b, err := yaml.Marshal(&out)
	if err != nil {
		return "", errors.Wrap(err, "marshal project yaml")
	}
	return string(b), nil
}

func (p *ProjectDB) ImportFromYAML(ctx context.Context, yamlContent string) error {
	if p.db == nil {
		return errors.New("project db not open")
	}

	var data projectYAML
	if err := yaml.Unmarshal([]byte(yamlContent), &data); err != nil {
		return errors.Wrap(err, "parse project yaml")
	}

	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return errors.Wrap(err, "begin tx")
	}
	defer func() { _ = tx.Rollback() }()

	// Clear existing data (mirror JS).
	for _, q := range []string{
		"DELETE FROM edges",
		"DELETE FROM nodes",
		"DELETE FROM project",
	} {
		if _, err := tx.ExecContext(ctx, q); err != nil {
			return errors.Wrap(err, "clear project db")
		}
	}

	if data.Project.Name != "" || data.Project.RootGoal != "" {
		stmt, err := tx.PrepareContext(ctx, "INSERT OR REPLACE INTO project (key, value) VALUES (?, ?)")
		if err != nil {
			return errors.Wrap(err, "prepare project meta upsert")
		}
		if _, err := stmt.ExecContext(ctx, "name", data.Project.Name); err != nil {
			_ = stmt.Close()
			return errors.Wrap(err, "set project name")
		}
		if _, err := stmt.ExecContext(ctx, "root_goal", data.Project.RootGoal); err != nil {
			_ = stmt.Close()
			return errors.Wrap(err, "set project root_goal")
		}
		_ = stmt.Close()
	}

	// Insert nodes.
	insertNodeStmt, err := tx.PrepareContext(ctx, `
INSERT INTO nodes (id, type, output, status, created_by, created_at, completed_at, parent_tactic, introduced_as, data)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
`)
	if err != nil {
		return errors.Wrap(err, "prepare insert node")
	}
	defer func() { _ = insertNodeStmt.Close() }()

	for id, n := range data.Nodes {
		status := n.Status
		if status == "" {
			status = "pending"
		}

		createdAt := time.Now().UTC().Format(time.RFC3339Nano)
		if n.CreatedAt != nil && *n.CreatedAt != "" {
			createdAt = *n.CreatedAt
		}

		var completedAt any
		if n.CompletedAt != nil && *n.CompletedAt != "" {
			completedAt = *n.CompletedAt
		} else {
			completedAt = nil
		}

		var dataJSON any
		if n.Data != nil {
			b, err := json.Marshal(n.Data)
			if err != nil {
				return errors.Wrap(err, "marshal node data")
			}
			dataJSON = string(b)
		} else {
			dataJSON = nil
		}

		if _, err := insertNodeStmt.ExecContext(
			ctx,
			id,
			n.Type,
			n.Output,
			status,
			n.CreatedBy,
			createdAt,
			completedAt,
			n.ParentTactic,
			n.IntroducedAs,
			dataJSON,
		); err != nil {
			return errors.Wrapf(err, "insert node %s", id)
		}
	}

	// Insert edges from dependencies.match
	insertEdgeStmt, err := tx.PrepareContext(ctx, "INSERT OR IGNORE INTO edges (source_node_id, target_node_id) VALUES (?, ?)")
	if err != nil {
		return errors.Wrap(err, "prepare insert edge")
	}
	defer func() { _ = insertEdgeStmt.Close() }()

	for targetID, n := range data.Nodes {
		if n.Dependencies == nil || len(n.Dependencies.Match) == 0 {
			continue
		}
		for _, sourceID := range n.Dependencies.Match {
			if _, err := insertEdgeStmt.ExecContext(ctx, sourceID, targetID); err != nil {
				return errors.Wrapf(err, "insert edge %s -> %s", sourceID, targetID)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return errors.Wrap(err, "commit tx")
	}
	return nil
}

// Debug helper: mostly useful for quick manual inspection.
func (p *ProjectDB) String() string {
	return fmt.Sprintf("ProjectDB(%s)", p.DBPath)
}
