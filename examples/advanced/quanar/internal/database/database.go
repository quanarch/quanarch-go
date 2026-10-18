// Package database is the quanar's one database connection. Quans never see it; they send operations.
package database

import (
	"database/sql"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite" // pure Go: no cgo, builds everywhere
)

// Open opens the database file, creating its folder.
func Open(file string) (*sql.DB, error) {
	if err := os.MkdirAll(filepath.Dir(file), 0o755); err != nil {
		return nil, err
	}
	// WAL with synchronous NORMAL: a crash of the process loses nothing; only a power cut may
	// lose the last commit. FULL would wait for the disk on every write.
	db, err := sql.Open("sqlite", "file:"+file+"?_pragma=journal_mode(WAL)&_pragma=synchronous(NORMAL)&_pragma=busy_timeout(3000)")
	if err != nil {
		return nil, err
	}
	// One connection: transactions and the writes around them stay in order.
	db.SetMaxOpenConns(1)
	return db, db.Ping()
}
