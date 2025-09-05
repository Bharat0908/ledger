package repo_test

import (
	"context"
	"testing"

	"github.com/Bharat0908/ledger/internal/repo"
	"github.com/google/uuid"
)

func TestPGRepo_CreateAccount(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		owner    string
		currency string
		initial  int64
		want     uuid.UUID
		wantErr  bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// TODO: construct the receiver type.
			var r repo.PGRepo
			got, gotErr := r.CreateAccount(context.Background(), tt.owner, tt.currency, tt.initial)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("CreateAccount() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("CreateAccount() succeeded unexpectedly")
			}
			// TODO: update the condition below to compare got with tt.want.
			if true {
				t.Errorf("CreateAccount() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPGRepo_GetAccount(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		id      uuid.UUID
		want    int64
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// TODO: construct the receiver type.
			var r repo.PGRepo
			got, gotErr := r.GetAccount(context.Background(), tt.id)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("GetAccount() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("GetAccount() succeeded unexpectedly")
			}
			// TODO: update the condition below to compare got with tt.want.
			if true {
				t.Errorf("GetAccount() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPGRepo_ApplyTransaction(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		accountID uuid.UUID
		typ       string
		amount    int64
		key       string
		want      int64
		wantErr   bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// TODO: construct the receiver type.
			var r repo.PGRepo
			got, gotErr := r.ApplyTransaction(context.Background(), tt.accountID, tt.typ, tt.amount, tt.key)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("ApplyTransaction() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("ApplyTransaction() succeeded unexpectedly")
			}
			// TODO: update the condition below to compare got with tt.want.
			if true {
				t.Errorf("ApplyTransaction() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPGRepo_ApplyTransfer(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		from    uuid.UUID
		to      uuid.UUID
		amount  int64
		key     string
		want    int64
		want2   int64
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// TODO: construct the receiver type.
			var r repo.PGRepo
			got, got2, gotErr := r.ApplyTransfer(context.Background(), tt.from, tt.to, tt.amount, tt.key)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("ApplyTransfer() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("ApplyTransfer() succeeded unexpectedly")
			}
			// TODO: update the condition below to compare got with tt.want.
			if true {
				t.Errorf("ApplyTransfer() = %v, want %v", got, tt.want)
			}
			if true {
				t.Errorf("ApplyTransfer() = %v, want %v", got2, tt.want2)
			}
		})
	}
}
