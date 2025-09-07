# Golang Banking Ledger

# Package content
This package contains a complete backend service for a banking ledger application:
- Go API + worker
- PostgreSQL (balances + idempotency)
- MongoDB (ledger)
- RabbitMQ (queue)
- Docker Compose and migrations
- Detailed OpenAPI spec and Postman collection
- Initialization scripts
- Test script

# Overview

A **Golang banking ledger**  application that is horizontally scalable, ACID-safe for core balance mutations using thread safe locking, and uses an **async queue** to process transactions under high load. It provides:

* **API Gateway** (Go, chi) for account creation and queuing transactions (idempotent).
* **Transaction Processor Worker** (Go) consuming from **RabbitMQ**, applying balance updates in **PostgreSQL** with row-level locks and idempotency dedupe, then writing ledger entries to **MongoDB**.
* **Docker Compose** to spin up Postgres, MongoDB, RabbitMQ, API, and Worker.
* **Testing**: unit (with mocks), integration (against ephemeral containers), and feature tests.

> Money values are handled as **integers of minor units (cents/paise)** to avoid floating-point errors.

## High-Level Architecture

```
Client → API (HTTP) ──publish──▶ RabbitMQ ──consume──▶ Worker
                                         │                │
                                         ▼                ▼
                                    Retry / DLQ      Postgres (balances)
                                                       +
                                                    MongoDB (ledger)
```


NOTE: Run `go mod tidy` locally to populate `go.sum` with accurate checksums before building.

## Testing Strategy

### Unit (Service/Repo with Mocks)

* Use `testify` for assertions and **interface-driven design** to mock repositories and queue.
* Examples:

  * **AccountService**: validates inputs; repo error surfaces.
  * **Tx flow**: ensures correct pub payload and idempotency key propagation.

### Integration

* Spin up **Postgres**, **Mongo**, **RabbitMQ** via `docker-compose`; run `go test -tags=integration` to:

  1. create account
  2. POST transaction (deposit/withdraw)
  3. wait for worker to process
  4. assert Postgres balance and Mongo ledger entry.
  5. initiate transer from one account to another 

---

## Prerequisites

1. Docker compose should be installed
2. Golang 1.22 version should be installed.

One can use the **init.sh** script to install the required utilities


## Quick Start

Load the golang environment using below command.
 * export PATH=$PATH:/usr/local/go/bin

1. `make up` (builds + starts Postgres, Mongo, RabbitMQ, API, Worker)
2. `make migrate`
3. Create account:

   ```bash
   curl -XPOST localhost:8080/v1/accounts -d '{"owner":"Alice","currency":"INR","initial_balance":10000}' -H 'Content-Type: application/json'
   ```
4. Enqueue deposit:

   ```bash
   curl -XPOST localhost:8080/v1/transactions \
     -H 'Content-Type: application/json' \
     -H 'Idempotency-Key: k-123' \
     -d '{"account_id":"<uuid>","type":"deposit","amount":5000}'
   ```
   Belwo are some examples
  a. Depositing 200...
  ```bash
  curl -s -X POST "localhost:8080/transactions" \
  -H 'Content-Type: application/json' \
  -H 'Idempotency-Key: tx-deposit-1' \
  -d '{"account_id":"ACCOUNT_ID","type":"deposit","amount":200}'
  ```
 b. Withdrawing 100...
  ```bash
  curl -s -X POST "localhost:8080/transactions" \
  -H 'Content-Type: application/json' \
  -H 'Idempotency-Key: tx-withdraw-1' \
  -d '{"account_id":"ACCOUNT_ID","type":"withdraw","amount":100}'
  ```
4. Transfer Betwwen accounts:
  ```bash
  curl -s -X POST "localhost:8080/transfers"   -H "Content-Type: application/json"   -H "Idempotency-Key: tx-withdraw-1"   -d '{
  "from_account_id":"100a8d3a-79ad-475c-bb5c-ac35910053dd",
  "to_account_id":"1547a2a8-a6bc-4f03-82ca-78e124a982d6",
  "amount":500,
  "idempotency_key":"transfer-1"
  }'
  ```

---

## Next Steps 

* Add **Kubernetes manifests** and **Helm chart** for orchestrated environment.
* Application can be customised to support **AWS Environment**.
  RabbitMQ - Amazon MQ
  MongoDB - DynamoDB/DocumentDB
  PostgresSQL - Amazon RDS/Aurora
  Docker-compose - Amazon ECS
  Kubernetes - Amazon EKS

