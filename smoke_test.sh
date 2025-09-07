#!/usr/bin/env bash
set -euo pipefail

API_URL="http://localhost:8080/v1"

echo " Starting services..."
#docker-compose up -d --build

echo "⏳ Waiting for API to be ready..."
until curl -s "$API_URL/accounts" >/dev/null 2>&1; do
  sleep 2
done

echo "✅ API is up!"

# 1. Create account
echo " Creating account..."
ACCOUNT_ID=$(curl -s -X POST "$API_URL/accounts" \
  -H "Content-Type: application/json" \
  -d '{"owner":"Alice","currency":"USD","initial_balance":1000}' \
  | jq -r '.id')

echo " Account created: $ACCOUNT_ID"

# 2. Deposit 200
echo " Depositing 200..."
curl -s -X POST "$API_URL/transactions" \
  -H "Content-Type: application/json" \
  -H "Idempotency-Key: tx-deposit-1" \
  -d "{\"account_id\":\"$ACCOUNT_ID\",\"type\":\"deposit\",\"amount\":200}"

# 3. Withdraw 100
echo " Withdrawing 100..."
curl -s -X POST "$API_URL/transactions" \
  -H "Content-Type: application/json" \
  -H "Idempotency-Key: tx-withdraw-1" \
  -d "{\"account_id\":\"$ACCOUNT_ID\",\"type\":\"withdraw\",\"amount\":100}"

# 4. Wait for worker to process
echo "⏳ Waiting for transactions to be processed..."
sleep 5

# 5. Fetch ledger
echo " Ledger entries:"
curl -s "$API_URL/accounts/$ACCOUNT_ID/ledger" | jq .

# 6. Verify balance
echo " Final balance (from DB):"
docker-compose exec -T postgres psql -U postgres -d ledger -c "SELECT balance FROM accounts WHERE id = '$ACCOUNT_ID';"

