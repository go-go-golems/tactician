package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"strings"

	"github.com/pkg/errors"
)

type TacticsDB struct {
	db *sql.DB
}

func NewTacticsDBFromDB(db *sql.DB) *TacticsDB {
	return &TacticsDB{db: db}
}

func (t *TacticsDB) InitSchema(ctx context.Context) error {
	if t.db == nil {
		return errors.New("tactics db not open")
	}

	const schemaSQL = `
CREATE TABLE IF NOT EXISTS tactics (
  id TEXT PRIMARY KEY,
  type TEXT NOT NULL,
  output TEXT NOT NULL,
  description TEXT,
  tags TEXT,
  data TEXT
);

CREATE TABLE IF NOT EXISTS tactic_dependencies (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  tactic_id TEXT NOT NULL,
  dependency_type TEXT NOT NULL,
  artifact_type TEXT NOT NULL,
  FOREIGN KEY (tactic_id) REFERENCES tactics(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS tactic_subtasks (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  tactic_id TEXT NOT NULL,
  subtask_id TEXT NOT NULL,
  output TEXT NOT NULL,
  type TEXT NOT NULL,
  depends_on TEXT,
  data TEXT,
  FOREIGN KEY (tactic_id) REFERENCES tactics(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_tactics_type ON tactics(type);
CREATE INDEX IF NOT EXISTS idx_tactic_deps ON tactic_dependencies(tactic_id);
CREATE INDEX IF NOT EXISTS idx_tactic_subtasks ON tactic_subtasks(tactic_id);
`
	_, err := t.db.ExecContext(ctx, schemaSQL)
	return errors.Wrap(err, "init tactics schema")
}

func (t *TacticsDB) AddTactic(ctx context.Context, tactic *Tactic) error {
	if t.db == nil {
		return errors.New("tactics db not open")
	}
	if tactic == nil {
		return errors.New("nil tactic")
	}

	var tags *string
	if len(tactic.Tags) > 0 {
		s := strings.Join(tactic.Tags, ",")
		tags = &s
	}

	var data *string
	if tactic.Data != nil {
		b, err := json.Marshal(tactic.Data)
		if err != nil {
			return errors.Wrap(err, "marshal tactic data")
		}
		s := string(b)
		data = &s
	}

	tx, err := t.db.BeginTx(ctx, nil)
	if err != nil {
		return errors.Wrap(err, "begin tx")
	}
	defer func() { _ = tx.Rollback() }()

	_, err = tx.ExecContext(ctx, `
INSERT OR REPLACE INTO tactics (id, type, output, description, tags, data)
VALUES (?, ?, ?, ?, ?, ?)
`, tactic.ID, tactic.Type, tactic.Output, nullIfEmpty(tactic.Description), tags, data)
	if err != nil {
		return errors.Wrap(err, "upsert tactic")
	}

	// Delete existing dependencies and subtasks (mirror JS behavior).
	if _, err := tx.ExecContext(ctx, "DELETE FROM tactic_dependencies WHERE tactic_id = ?", tactic.ID); err != nil {
		return errors.Wrap(err, "delete tactic dependencies")
	}
	if _, err := tx.ExecContext(ctx, "DELETE FROM tactic_subtasks WHERE tactic_id = ?", tactic.ID); err != nil {
		return errors.Wrap(err, "delete tactic subtasks")
	}

	depStmt, err := tx.PrepareContext(ctx, `
INSERT INTO tactic_dependencies (tactic_id, dependency_type, artifact_type)
VALUES (?, ?, ?)
`)
	if err != nil {
		return errors.Wrap(err, "prepare dep insert")
	}
	defer func() { _ = depStmt.Close() }()

	for _, a := range tactic.Match {
		if _, err := depStmt.ExecContext(ctx, tactic.ID, "match", a); err != nil {
			return errors.Wrap(err, "insert match dependency")
		}
	}
	for _, a := range tactic.Premises {
		if _, err := depStmt.ExecContext(ctx, tactic.ID, "premise", a); err != nil {
			return errors.Wrap(err, "insert premise dependency")
		}
	}

	if len(tactic.Subtasks) > 0 {
		subStmt, err := tx.PrepareContext(ctx, `
INSERT INTO tactic_subtasks (tactic_id, subtask_id, output, type, depends_on, data)
VALUES (?, ?, ?, ?, ?, ?)
`)
		if err != nil {
			return errors.Wrap(err, "prepare subtask insert")
		}
		defer func() { _ = subStmt.Close() }()

		for _, st := range tactic.Subtasks {
			var dependsOn *string
			if len(st.DependsOn) > 0 {
				s := strings.Join(st.DependsOn, ",")
				dependsOn = &s
			}

			var stData *string
			if st.Data != nil {
				b, err := json.Marshal(st.Data)
				if err != nil {
					return errors.Wrap(err, "marshal subtask data")
				}
				s := string(b)
				stData = &s
			}

			if _, err := subStmt.ExecContext(ctx, tactic.ID, st.ID, st.Output, st.Type, dependsOn, stData); err != nil {
				return errors.Wrap(err, "insert subtask")
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return errors.Wrap(err, "commit tx")
	}
	return nil
}

func (t *TacticsDB) GetTactic(ctx context.Context, id string) (*Tactic, error) {
	if t.db == nil {
		return nil, errors.New("tactics db not open")
	}

	row := t.db.QueryRowContext(ctx, "SELECT id, type, output, description, tags, data FROM tactics WHERE id = ?", id)
	var tactic Tactic
	var description, tags, data sql.NullString
	if err := row.Scan(&tactic.ID, &tactic.Type, &tactic.Output, &description, &tags, &data); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, errors.Wrap(err, "scan tactic")
	}
	if description.Valid {
		tactic.Description = description.String
	}
	if tags.Valid && tags.String != "" {
		tactic.Tags = strings.Split(tags.String, ",")
	}
	if data.Valid && data.String != "" {
		var m map[string]interface{}
		if err := json.Unmarshal([]byte(data.String), &m); err != nil {
			return nil, errors.Wrap(err, "unmarshal tactic data")
		}
		tactic.Data = m
	}

	// Dependencies
	depRows, err := t.db.QueryContext(ctx, "SELECT dependency_type, artifact_type FROM tactic_dependencies WHERE tactic_id = ?", id)
	if err != nil {
		return nil, errors.Wrap(err, "select tactic deps")
	}
	defer func() { _ = depRows.Close() }()

	for depRows.Next() {
		var depType, artifact string
		if err := depRows.Scan(&depType, &artifact); err != nil {
			return nil, errors.Wrap(err, "scan tactic dep")
		}
		switch depType {
		case "match":
			tactic.Match = append(tactic.Match, artifact)
		case "premise":
			tactic.Premises = append(tactic.Premises, artifact)
		}
	}
	if err := depRows.Err(); err != nil {
		return nil, errors.Wrap(err, "iterate tactic deps")
	}

	// Subtasks
	subRows, err := t.db.QueryContext(ctx, "SELECT subtask_id, output, type, depends_on, data FROM tactic_subtasks WHERE tactic_id = ? ORDER BY id", id)
	if err != nil {
		return nil, errors.Wrap(err, "select subtasks")
	}
	defer func() { _ = subRows.Close() }()

	for subRows.Next() {
		var st TacticSubtask
		var dependsOn, stData sql.NullString
		if err := subRows.Scan(&st.ID, &st.Output, &st.Type, &dependsOn, &stData); err != nil {
			return nil, errors.Wrap(err, "scan subtask")
		}
		if dependsOn.Valid && dependsOn.String != "" {
			st.DependsOn = strings.Split(dependsOn.String, ",")
		}
		if stData.Valid && stData.String != "" {
			var m map[string]interface{}
			if err := json.Unmarshal([]byte(stData.String), &m); err != nil {
				return nil, errors.Wrap(err, "unmarshal subtask data")
			}
			st.Data = m
		}
		tactic.Subtasks = append(tactic.Subtasks, st)
	}
	if err := subRows.Err(); err != nil {
		return nil, errors.Wrap(err, "iterate subtasks")
	}

	return &tactic, nil
}

func (t *TacticsDB) GetAllTactics(ctx context.Context) ([]*Tactic, error) {
	if t.db == nil {
		return nil, errors.New("tactics db not open")
	}

	rows, err := t.db.QueryContext(ctx, "SELECT id FROM tactics")
	if err != nil {
		return nil, errors.Wrap(err, "select tactic ids")
	}
	defer func() { _ = rows.Close() }()

	var ret []*Tactic
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, errors.Wrap(err, "scan tactic id")
		}
		tactic, err := t.GetTactic(ctx, id)
		if err != nil {
			return nil, err
		}
		if tactic != nil {
			ret = append(ret, tactic)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "iterate tactic ids")
	}
	return ret, nil
}

// SearchTactics provides a minimal filter layer similar to JS searchTactics(filters).
// Ranking logic lives in the command/helper layer.
func (t *TacticsDB) SearchTactics(ctx context.Context, typeFilter string, tags []string, keywords []string) ([]*Tactic, error) {
	if t.db == nil {
		return nil, errors.New("tactics db not open")
	}

	query := "SELECT DISTINCT t.id FROM tactics t"
	var conditions []string
	var params []any

	if typeFilter != "" {
		conditions = append(conditions, "t.type = ?")
		params = append(params, typeFilter)
	}

	if len(tags) > 0 {
		var tagConds []string
		for range tags {
			tagConds = append(tagConds, "t.tags LIKE ?")
		}
		conditions = append(conditions, "("+strings.Join(tagConds, " OR ")+")")
		for _, tag := range tags {
			params = append(params, "%"+tag+"%")
		}
	}

	if len(keywords) > 0 {
		var kwConds []string
		for range keywords {
			kwConds = append(kwConds, "(t.id LIKE ? OR t.description LIKE ? OR t.tags LIKE ?)")
		}
		conditions = append(conditions, "("+strings.Join(kwConds, " OR ")+")")
		for _, kw := range keywords {
			pat := "%" + kw + "%"
			params = append(params, pat, pat, pat)
		}
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	rows, err := t.db.QueryContext(ctx, query, params...)
	if err != nil {
		return nil, errors.Wrap(err, "search tactics")
	}
	defer func() { _ = rows.Close() }()

	var ret []*Tactic
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, errors.Wrap(err, "scan tactic id")
		}
		tactic, err := t.GetTactic(ctx, id)
		if err != nil {
			return nil, err
		}
		if tactic != nil {
			ret = append(ret, tactic)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "iterate tactics search results")
	}

	return ret, nil
}

func nullIfEmpty(s string) any {
	if s == "" {
		return nil
	}
	return s
}
