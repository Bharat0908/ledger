#!/usr/bin/env bash
docker-compose exec -T postgres psql -U postgres -d ledger -c "CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\";"
docker-compose exec -T postgres psql -U postgres -d ledger -c "CREATE TABLE IF NOT EXISTS accounts (
 id UUID PRIMARY KEY,
 owner TEXT NOT NULL,
 currency TEXT NOT NULL,
 balance BIGINT NOT NULL,
 created_at TIMESTAMPTZ NOT NULL DEFAULT now()
 );
 "
#docker-compose exec -T postgres psql -U postgres -d ledger -c "SELECT balance FROM accounts;"
docker-compose exec -T postgres psql -U postgres -d ledger -c "CREATE TABLE IF NOT EXISTS processed_messages (
 idempotency_key TEXT PRIMARY KEY,
 account_id UUID NOT NULL,
 type TEXT NOT NULL,
 amount BIGINT NOT NULL,
 processed_at TIMESTAMPTZ NOT NULL DEFAULT now()
 );"
docker-compose exec -T postgres psql -U postgres -d ledger -c "CREATE INDEX IF NOT EXISTS idx_accounts_owner ON accounts(owner);"


