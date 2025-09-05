package queue

import "time"

// TxMessage represents a transaction message to be processed in the queue.
// It contains information about the account, transaction type, amount,
// idempotency key, and the creation timestamp.
type TxMessage struct {
	AccountID string    `json:"account_id"`
	Type      string    `json:"type"`
	Amount    int64     `json:"amount"`
	Key       string    `json:"idempotency_key"`
	CreatedAt time.Time `json:"created_at"`
}

// TransferMessage represents a message containing the details of a transfer operation
// between two accounts. It includes the source and destination account IDs, the amount
// to be transferred, an idempotency key to ensure operation uniqueness, and the timestamp
// when the message was created.
type TransferMessage struct {
	FromAccountID string    `json:"from_account_id"`
	ToAccountID   string    `json:"to_account_id"`
	Amount        int64     `json:"amount"`
	Key           string    `json:"idempotency_key"`
	CreatedAt     time.Time `json:"created_at"`
}
