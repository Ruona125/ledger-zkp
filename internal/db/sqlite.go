package db

import (
	"database/sql"

	_ "modernc.org/sqlite"
)

func Open() *sql.DB {
	db, err := sql.Open("sqlite", "file:ledger.db?cache=shared&mode=rwc")
	if err != nil {
		panic(err)
	}
	if err := Migrate(db); err != nil {
		panic(err)
	}
	return db
}
