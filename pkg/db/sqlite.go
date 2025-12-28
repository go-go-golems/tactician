package db

import (
	"context"
	"database/sql"

	_ "modernc.org/sqlite"

	"github.com/pkg/errors"
)

func openSQLite(ctx context.Context, path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, errors.Wrap(err, "open sqlite")
	}

	// Make sure the DB is reachable early.
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, errors.Wrap(err, "ping sqlite")
	}

	// Mirror JS behavior: WAL.
	if _, err := db.ExecContext(ctx, "PRAGMA journal_mode = WAL;"); err != nil {
		_ = db.Close()
		return nil, errors.Wrap(err, "set journal_mode=WAL")
	}

	return db, nil
}
