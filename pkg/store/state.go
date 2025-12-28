package store

import (
	"context"
	"database/sql"

	"github.com/go-go-golems/tactician/pkg/db"
	"github.com/pkg/errors"
)

type State struct {
	Dir string

	SQL     *sql.DB
	Project *db.ProjectDB

	// TODO(manuel): Add TacticsDB once implemented.
	Dirty bool
}

func Load(ctx context.Context, tacticianDir string) (*State, error) {
	sqlDB, err := db.OpenSQLiteMemory(ctx)
	if err != nil {
		return nil, err
	}

	projectDB := db.NewProjectDBFromDB(sqlDB)
	if err := projectDB.InitSchema(ctx); err != nil {
		_ = sqlDB.Close()
		return nil, err
	}

	// TODO(manuel): Import YAML from tacticianDir into the in-memory DB.
	return &State{
		Dir:     tacticianDir,
		SQL:     sqlDB,
		Project: projectDB,
		Dirty:   false,
	}, nil
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
	_ = ctx
	if s == nil {
		return errors.New("nil state")
	}
	// TODO(manuel): Export in-memory DB back to YAML in s.Dir.
	return nil
}
