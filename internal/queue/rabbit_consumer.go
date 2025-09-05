package queue

import (
	"context"
	"encoding/json"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// Interfaces for dependency injection
type BalanceApplier interface {
	Apply(ctx context.Context, accID, typ string, amount int64, key string) (int64, error)
	ApplyTransfer(ctx context.Context, from, to string, amount int64, key string) (fromAfter, toAfter int64, err error)
}

// LedgerWriter defines methods for writing ledger entries, including single account operations
// and transfers between accounts. Implementations are responsible for persisting these entries
// with the provided context, account identifiers, transaction details, and timestamps.
type LedgerWriter interface {
	Write(ctx context.Context, accID, typ string, amount, balanceAfter int64, key string, at time.Time) error
	WriteTransfer(ctx context.Context, from, to string, amount, fromAfter, toAfter int64, key string, at time.Time) error
}

// Consumer represents a RabbitMQ consumer that processes messages from a specified queue.
// It holds a reference to the AMQP channel, the queue name, a BalanceApplier for applying balance changes,
// and a LedgerWriter for recording ledger entries.
type Consumer struct {
	Ch      *amqp.Channel
	Queue   string
	Applier BalanceApplier
	Ledger  LedgerWriter
}

// Start begins consuming messages from the configured RabbitMQ queue and processes them.
// It listens for messages of type TxMessage or TransferMessage, applies the corresponding
// ledger operations, and acknowledges or negatively acknowledges messages based on the
// processing result. The method runs until the provided context is canceled, at which point
// it returns. If an error occurs during queue consumption setup, it is returned immediately.
//
// Parameters:
//   - ctx: Context for cancellation and timeout control.
//
// Returns:
//   - error: An error if queue consumption setup fails or if the context is canceled.
func (c *Consumer) Start(ctx context.Context) error {
	deliveries, err := c.Ch.Consume(
		c.Queue,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case d := <-deliveries:
			var m TxMessage
			if err := json.Unmarshal(d.Body, &m); err == nil && m.AccountID != "" {
				bal, err := c.Applier.Apply(ctx, m.AccountID, m.Type, m.Amount, m.Key)
				if err != nil {
					// requeue (true) for transient errors; if permanent, consider DLQ
					d.Nack(false, true)
					continue
				}
				if err := c.Ledger.Write(ctx, m.AccountID, m.Type, m.Amount, bal, m.Key, m.CreatedAt); err != nil {
					d.Nack(false, true)
					continue
				}
				d.Ack(false)
				continue
			}

			// Try as transfer
			var t TransferMessage
			if err := json.Unmarshal(d.Body, &t); err == nil && t.FromAccountID != "" && t.ToAccountID != "" {
				fromAfter, toAfter, err := c.Applier.ApplyTransfer(ctx, t.FromAccountID, t.ToAccountID, t.Amount, t.Key)
				if err != nil {
					d.Nack(false, true)
					continue
				}
				if err := c.Ledger.WriteTransfer(ctx, t.FromAccountID, t.ToAccountID, t.Amount, fromAfter, toAfter, t.Key, t.CreatedAt); err != nil {
					d.Nack(false, true)
					continue
				}
				d.Ack(false)
				continue
			}

			// Unknown payload
			d.Nack(false, false)
		}
	}
}
