package repo

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PGRepo struct{ DB *pgxpool.Pool }

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

func (r *PGRepo) GetAccount(ctx context.Context, id uuid.UUID) (int64, error) {
	var bal int64
	if err := r.DB.QueryRow(ctx, `SELECT balance FROM accounts WHERE id=$1`, id).Scan(&bal); err != nil {
		return 0, err
	}
	return bal, nil
}

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
