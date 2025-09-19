package ledger

import (
	"context"
	"database/sql"
	"errors"
)

type Service struct{ db *sql.DB }

func NewService(db *sql.DB) *Service { return &Service{db: db} }

var ErrInsufficientFunds = errors.New("insufficient funds")

// ApplyTx ensures idempotency + non-negative balance.
func (s *Service) ApplyTx(ctx context.Context, accountID, idempotencyKey, entryID string, amount int64) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// idempotency guard
	if _, err := tx.ExecContext(ctx, `INSERT INTO idempotency_keys(key) VALUES(?)`, idempotencyKey); err != nil {
		return nil // duplicate -> treat as success
	}

	// if debit, check available balance
	if amount < 0 {
		var bal int64
		if err := tx.QueryRowContext(ctx,
			`SELECT COALESCE(SUM(amount),0) FROM ledger_entries WHERE account_id=?`,
			accountID).Scan(&bal); err != nil {
			return err
		}
		if bal+amount < 0 {
			return ErrInsufficientFunds
		}
	}

	if _, err := tx.ExecContext(ctx,
		`INSERT INTO ledger_entries(id,account_id,amount) VALUES(?,?,?)`,
		entryID, accountID, amount); err != nil {
		return err
	}

	return tx.Commit()
}
