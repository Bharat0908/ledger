package queue

import "time"

type TxMessage struct {
	AccountID string    `json:"account_id"`
	Type      string    `json:"type"`
	Amount    int64     `json:"amount"`
	Key       string    `json:"idempotency_key"`
	CreatedAt time.Time `json:"created_at"`
}

type TransferMessage struct {
	FromAccountID string    `json:"from_account_id"`
	ToAccountID   string    `json:"to_account_id"`
	Amount        int64     `json:"amount"`
	Key           string    `json:"idempotency_key"`
	CreatedAt     time.Time `json:"created_at"`
}
