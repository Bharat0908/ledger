#!/usr/bin/env bash
set -euo pipefail

API_BASEURL="http://localhost:8080"
API_URL="$API_BASEURL/v1"

echo "🚀 Starting services..."
#docker-compose up -d --build

echo "⏳ Waiting for API to be ready..."
until curl -sSf "$API_BASEURL/healthz" >/dev/null 2>&1; do
  sleep 2
done

echo "✅ API is up!"

# 1. Create account
echo "📂 Creating account..."
ACCOUNT_ID=$(curl -s -X POST "$API_URL/accounts"   -H "Content-Type: application/json"   -d '{"owner":"Alice","currency":"USD","initial_balance":1000}'   | jq -r '.id')

if [ -z "$ACCOUNT_ID" ] || [ "$ACCOUNT_ID" = "null" ]; then
  echo "Failed to create account or parse ID"
  exit 1
fi

echo "👉 Account created: $ACCOUNT_ID"

# 2. Deposit 200
echo "💰 Depositing 200..."
curl -s -X POST "$API_URL/transactions"   -H "Content-Type: application/json"   -H "Idempotency-Key: tx-deposit-1"   -d "{\"account_id\":\"$ACCOUNT_ID\",\"type\":\"deposit\",\"amount\":200}"

# 3. Withdraw 100
echo "🏧 Withdrawing 100..."
curl -s -X POST "$API_URL/transactions"   -H "Content-Type: application/json"   -H "Idempotency-Key: tx-withdraw-1"   -d "{\"account_id\":\"$ACCOUNT_ID\",\"type\":\"withdraw\",\"amount\":100}"

# 4. Wait for worker to process
echo "⏳ Waiting for transactions to be processed..."
sleep 6

# 5. Fetch ledger
echo "📜 Ledger entries:"
curl -s "$API_URL/accounts/$ACCOUNT_ID/ledger" | jq .

curl -s -X POST "$API_URL/transfers"   -H "Content-Type: application/json"   -H "Idempotency-Key: tx-withdraw-1"   -d '{
  "from_account_id":"100a8d3a-79ad-475c-bb5c-ac35910053dd",
  "to_account_id":"1547a2a8-a6bc-4f03-82ca-78e124a982d6",
  "amount":500,
  "idempotency_key":"transfer-1"
}'


# 6. Verify balance
echo "💳 Final balance (from DB):"
docker-compose exec -T postgres psql -U postgres -d ledger -c "SELECT balance FROM accounts WHERE id = '$ACCOUNT_ID';"
