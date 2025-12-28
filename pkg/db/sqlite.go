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

// OpenSQLiteMemory opens a shared in-memory sqlite database.
// This is the only runtime mode we want for tactician: YAML on disk, sqlite in memory.
func OpenSQLiteMemory(ctx context.Context) (*sql.DB, error) {
	// "file::memory:?cache=shared" allows the driver to share an in-memory database across connections.
	// We still treat the returned *sql.DB as a single handle for the process.
	return openSQLite(ctx, "file::memory:?cache=shared")
}
