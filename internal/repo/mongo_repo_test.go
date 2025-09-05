package repo_test

import (
	"context"
	"testing"
	"time"

	"github.com/Bharat0908/ledger/internal/repo"
	"github.com/google/uuid"
)

func TestMongoRepo_InsertLedger(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		accountID    uuid.UUID
		typ          string
		amount       int64
		balanceAfter int64
		key          string
		at           time.Time
		wantErr      bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// TODO: construct the receiver type.
			var m repo.MongoRepo
			gotErr := m.InsertLedger(context.Background(), tt.accountID, tt.typ, tt.amount, tt.balanceAfter, tt.key, tt.at)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("InsertLedger() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("InsertLedger() succeeded unexpectedly")
			}
		})
	}
}

func TestMongoRepo_InsertTransferLedger(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		from      uuid.UUID
		to        uuid.UUID
		amount    int64
		fromAfter int64
		toAfter   int64
		key       string
		at        time.Time
		wantErr   bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// TODO: construct the receiver type.
			var m repo.MongoRepo
			gotErr := m.InsertTransferLedger(context.Background(), tt.from, tt.to, tt.amount, tt.fromAfter, tt.toAfter, tt.key, tt.at)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("InsertTransferLedger() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("InsertTransferLedger() succeeded unexpectedly")
			}
		})
	}
}

func TestMongoRepo_GetTransactions(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		accountID string
		limit     int
		want      []map[string]interface{}
		wantErr   bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// TODO: construct the receiver type.
			var m repo.MongoRepo
			got, gotErr := m.GetTransactions(context.Background(), tt.accountID, tt.limit)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("GetTransactions() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("GetTransactions() succeeded unexpectedly")
			}
			// TODO: update the condition below to compare got with tt.want.
			if true {
				t.Errorf("GetTransactions() = %v, want %v", got, tt.want)
			}
		})
	}
}
