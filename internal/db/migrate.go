package db

import "database/sql"

func Migrate(db *sql.DB) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS accounts (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			public_hash TEXT NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS ledger_entries (
			id TEXT PRIMARY KEY,
			account_id TEXT NOT NULL,
			amount INTEGER NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);`,
		`CREATE TABLE IF NOT EXISTS idempotency_keys (
			key TEXT PRIMARY KEY
		);`,
	}
	for _, s := range stmts {
		if _, err := db.Exec(s); err != nil {
			return err
		}
	}
	return nil
}
