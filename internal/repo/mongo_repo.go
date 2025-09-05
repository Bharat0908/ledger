// MongoRepo provides methods to interact with a MongoDB collection for ledger operations.
// It embeds a mongo.Collection to perform database operations such as inserting ledger entries
// and retrieving transaction histories.
//
// InsertLedger inserts a single ledger entry for a given account, specifying the type, amount,
// resulting balance, idempotency key, and creation timestamp.
//
// InsertTransferLedger inserts two ledger entries in a single operation to represent a transfer
// between two accounts: a debit from the sender and a credit to the receiver, each with their
// respective resulting balances, idempotency key, and timestamp.
//
// GetTransactions retrieves a limited number of recent transactions for a given account ID,
// sorted by creation time in descending order. It returns the transactions as a slice of maps.
package repo

import (
	"context"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoRepo provides methods to interact with a MongoDB collection.
// It embeds a mongo.Collection to perform database operations.
type MongoRepo struct{ C *mongo.Collection }

func (m *MongoRepo) InsertLedger(ctx context.Context, accountID uuid.UUID, typ string, amount, balanceAfter int64, key string, at time.Time) error {
	_, err := m.C.InsertOne(ctx, bson.M{
		"account_id":      accountID.String(),
		"type":            typ,
		"amount":          amount,
		"balance_after":   balanceAfter,
		"idempotency_key": key,
		"created_at":      at,
	})
	return err
}

// InsertTransferLedger inserts two ledger entries representing a transfer between accounts.
// It creates a debit entry for the source account and a credit entry for the destination account,
// both associated with the same idempotency key and timestamp. The operation is performed as a single
// insert of two documents to ensure atomicity at the database level.
//
// Parameters:
//   - ctx: Context for controlling cancellation and deadlines.
//   - from: UUID of the source account.
//   - to: UUID of the destination account.
//   - amount: Amount to transfer.
//   - fromAfter: Balance of the source account after the transfer.
//   - toAfter: Balance of the destination account after the transfer.
//   - key: Idempotency key to prevent duplicate transfers.
//   - at: Timestamp of the transfer.
//
// Returns:
//   - error: Non-nil if the insert operation fails.
func (m *MongoRepo) InsertTransferLedger(ctx context.Context, from, to uuid.UUID, amount, fromAfter, toAfter int64, key string, at time.Time) error {
	// insert two documents in a single operation
	docs := []interface{}{
		bson.M{"account_id": from.String(), "type": "transfer_debit", "amount": -amount, "balance_after": fromAfter, "idempotency_key": key, "created_at": at},
		bson.M{"account_id": to.String(), "type": "transfer_credit", "amount": amount, "balance_after": toAfter, "idempotency_key": key, "created_at": at},
	}
	_, err := m.C.InsertMany(ctx, docs)
	return err
}

// GetTransactions retrieves a list of transactions for the specified account ID from the MongoDB collection.
// The transactions are sorted by the "created_at" field in descending order and limited to the specified number.
// It returns a slice of maps representing the transactions and an error if the operation fails.
//
// Parameters:
//   - ctx: The context for controlling cancellation and timeouts.
//   - accountID: The ID of the account whose transactions are to be retrieved.
//   - limit: The maximum number of transactions to return.
//
// Returns:
//   - []map[string]interface{}: A slice of transactions represented as maps.
//   - error: An error if the retrieval fails, otherwise nil.
func (m *MongoRepo) GetTransactions(ctx context.Context, accountID string, limit int) ([]map[string]interface{}, error) {
	filter := bson.M{"account_id": accountID}
	opts := options.Find().SetSort(bson.D{{"created_at", -1}}).SetLimit(int64(limit))
	cur, err := m.C.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	var out []map[string]interface{}
	for cur.Next(ctx) {
		var doc map[string]interface{}
		if err := cur.Decode(&doc); err != nil {
			return nil, err
		}
		out = append(out, doc)
	}
	return out, nil
}
