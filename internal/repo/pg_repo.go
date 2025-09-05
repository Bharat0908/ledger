package repo

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PGRepo provides methods to interact with a PostgreSQL database using a pgx connection pool.
//
// Methods:
//
//   - CreateAccount(ctx, owner, currency, initial):
//     Creates a new account with the specified owner, currency, and initial balance.
//     Returns the generated account UUID or an error.
//
//   - GetAccount(ctx, id):
//     Retrieves the balance of the account with the given UUID.
//     Returns the balance or an error.
//
//   - ApplyTransaction(ctx, accountID, typ, amount, key):
//     Applies a deposit or withdrawal transaction to the specified account, using an idempotency key to ensure the operation is not repeated.
//     Returns the new balance or an error.
//
//   - ApplyTransfer(ctx, from, to, amount, key):
//     Transfers the specified amount from one account to another, using an idempotency key to ensure the operation is not repeated.
//     Returns the new balances of both accounts or an error.
type PGRepo struct{ DB *pgxpool.Pool }

// CreateAccount creates a new account in the database with the specified owner, currency, and initial balance.
// It generates a new UUID for the account, inserts the account record into the "accounts" table within a transaction,
// and returns the generated account UUID upon success. If any error occurs during the process, it returns uuid.Nil and the error.
// The operation is performed within the provided context for cancellation and timeout control.
func (r *PGRepo) CreateAccount(ctx context.Context, owner, currency string, initial int64) (uuid.UUID, error) {
	id := uuid.New()
	tx, err := r.DB.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return uuid.Nil, err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `INSERT INTO accounts(id, owner, currency, balance, created_at) VALUES($1,$2,$3,$4,$5)`, id, owner, currency, initial, time.Now()); err != nil {
		return uuid.Nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return uuid.Nil, err
	}
	return id, nil
}

// GetAccount retrieves the balance of an account with the specified UUID from the database.
// It returns the account balance as an int64 and an error if the query fails or the account does not exist.
//
// Parameters:
//
//	ctx - The context for controlling cancellation and timeouts.
//	id  - The UUID of the account to retrieve.
//
// Returns:
//
//	int64 - The balance of the account.
//	error - An error if the account could not be retrieved.
func (r *PGRepo) GetAccount(ctx context.Context, id uuid.UUID) (int64, error) {
	var bal int64
	if err := r.DB.QueryRow(ctx, `SELECT balance FROM accounts WHERE id=$1`, id).Scan(&bal); err != nil {
		return 0, err
	}
	return bal, nil
}

// ApplyTransaction applies a deposit or withdrawal transaction to the specified account in a transactional manner.
// It ensures idempotency using the provided key, so duplicate requests with the same key will not result in double processing.
// The function locks the account row for update, checks for sufficient funds on withdrawal, updates the balance,
// and records the processed transaction. Returns the resulting balance after the transaction or an error.
//
// Parameters:
//
//	ctx       - context for cancellation and timeout control
//	accountID - UUID of the account to apply the transaction to
//	typ       - transaction type: "deposit" or "withdraw"
//	amount    - amount to deposit or withdraw
//	key       - idempotency key to ensure transaction uniqueness
//
// Returns:
//
//	balanceAfter - the account balance after the transaction
//	err          - error if the transaction failed or was invalid
func (r *PGRepo) ApplyTransaction(ctx context.Context, accountID uuid.UUID, typ string, amount int64, key string) (balanceAfter int64, err error) {
	tx, err := r.DB.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)

	// idempotency check
	var existing string
	err = tx.QueryRow(ctx, `SELECT idempotency_key FROM processed_messages WHERE idempotency_key=$1`, key).Scan(&existing)
	if err == nil {
		var bal int64
		if err := tx.QueryRow(ctx, `SELECT balance FROM accounts WHERE id=$1`, accountID).Scan(&bal); err != nil {
			return 0, err
		}
		return bal, tx.Commit(ctx)
	}

	var balance int64
	if err = tx.QueryRow(ctx, `SELECT balance FROM accounts WHERE id=$1 FOR UPDATE`, accountID).Scan(&balance); err != nil {
		return 0, err
	}

	switch typ {
	case "deposit":
		balance += amount
	case "withdraw":
		if balance < amount {
			return 0, errors.New("insufficient_funds")
		}
		balance -= amount
	default:
		return 0, errors.New("invalid_type")
	}

	if _, err = tx.Exec(ctx, `UPDATE accounts SET balance=$1 WHERE id=$2`, balance, accountID); err != nil {
		return 0, err
	}
	if _, err = tx.Exec(ctx, `INSERT INTO processed_messages(idempotency_key,account_id,type,amount,processed_at) VALUES($1,$2,$3,$4,$5)`, key, accountID, typ, amount, time.Now()); err != nil {
		return 0, err
	}
	if err = tx.Commit(ctx); err != nil {
		return 0, err
	}
	return balance, nil
}

// ApplyTransfer performs a transfer of the specified amount from one account to another within a database transaction.
// It ensures idempotency using the provided key, so repeated calls with the same key will not result in duplicate transfers.
// The function locks both accounts to prevent race conditions and deadlocks, and checks for sufficient funds before proceeding.
// On success, it returns the updated balances of the source and destination accounts.
// If the transfer has already been processed (as determined by the idempotency key), it returns the current balances without applying the transfer.
// Returns an error if the transaction fails, the accounts cannot be locked, or there are insufficient funds.
func (r *PGRepo) ApplyTransfer(ctx context.Context, from, to uuid.UUID, amount int64, key string) (fromAfter, toAfter int64, err error) {
	tx, err := r.DB.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return 0, 0, err
	}
	defer tx.Rollback(ctx)

	var existing string
	if err = tx.QueryRow(ctx, `SELECT idempotency_key FROM processed_messages WHERE idempotency_key=$1`, key).Scan(&existing); err == nil {
		// already processed
		var fb, tb int64
		if err := tx.QueryRow(ctx, `SELECT balance FROM accounts WHERE id=$1`, from).Scan(&fb); err != nil {
			return 0, 0, err
		}
		if err := tx.QueryRow(ctx, `SELECT balance FROM accounts WHERE id=$1`, to).Scan(&tb); err != nil {
			return 0, 0, err
		}
		return fb, tb, tx.Commit(ctx)
	}

	// order to avoid deadlocks
	first, second := from, to
	if to.String() < from.String() {
		first, second = to, from
	}

	rows, err := tx.Query(ctx, `SELECT id, balance FROM accounts WHERE id IN ($1,$2) FOR UPDATE`, first, second)
	if err != nil {
		return 0, 0, err
	}
	defer rows.Close()
	balances := map[string]int64{}
	for rows.Next() {
		var id uuid.UUID
		var bal int64
		if err := rows.Scan(&id, &bal); err != nil {
			return 0, 0, err
		}
		balances[id.String()] = bal
	}

	fromBal := balances[from.String()]
	toBal := balances[to.String()]
	if fromBal < amount {
		return 0, 0, errors.New("insufficient_funds")
	}
	fromBal -= amount
	toBal += amount

	if _, err := tx.Exec(ctx, `UPDATE accounts SET balance=$1 WHERE id=$2`, fromBal, from); err != nil {
		return 0, 0, err
	}
	if _, err := tx.Exec(ctx, `UPDATE accounts SET balance=$1 WHERE id=$2`, toBal, to); err != nil {
		return 0, 0, err
	}

	if _, err := tx.Exec(ctx, `INSERT INTO processed_messages(idempotency_key,account_id,type,amount,processed_at) VALUES($1,$2,$3,$4,$5)`, key, from, "transfer", amount, time.Now()); err != nil {
		return 0, 0, err
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, 0, err
	}
	return fromBal, toBal, nil
}
