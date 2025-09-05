package repo

import (
	"context"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

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

func (m *MongoRepo) InsertTransferLedger(ctx context.Context, from, to uuid.UUID, amount, fromAfter, toAfter int64, key string, at time.Time) error {
	// insert two documents in a single operation
	docs := []interface{}{
		bson.M{"account_id": from.String(), "type": "transfer_debit", "amount": -amount, "balance_after": fromAfter, "idempotency_key": key, "created_at": at},
		bson.M{"account_id": to.String(), "type": "transfer_credit", "amount": amount, "balance_after": toAfter, "idempotency_key": key, "created_at": at},
	}
	_, err := m.C.InsertMany(ctx, docs)
	return err
}

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
