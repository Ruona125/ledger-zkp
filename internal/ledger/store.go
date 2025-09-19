package ledger

import (
	"context"
	"database/sql"
)

type Store struct{ db *sql.DB }

func NewStore(db *sql.DB) *Store { return &Store{db: db} }

func (s *Store) CreateAccount(ctx context.Context, id, name, publicHash string) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO accounts(id,name,public_hash) VALUES(?,?,?)`,
		id, name, publicHash)
	return err
}

func (s *Store) Balance(ctx context.Context, accountID string) (int64, error) {
	var sum int64
	_ = s.db.QueryRowContext(ctx,
		`SELECT COALESCE(SUM(amount),0) FROM ledger_entries WHERE account_id=?`,
		accountID).Scan(&sum)
	return sum, nil
}
