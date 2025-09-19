package events

type TxCreated struct {
	EventID        string `json:"event_id"`
	AccountID      string `json:"account_id"`
	Amount         int64  `json:"amount"`
	IdempotencyKey string `json:"idempotency_key"`
}

const SubjectTxCreated = "tx.created"
