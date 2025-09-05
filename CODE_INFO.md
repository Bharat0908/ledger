# Ledger Service — Code Documentation

This repository implements a distributed ledger system using Go, PostgreSQL, MongoDB, and RabbitMQ. Below is an overview and documentation for the main code files.


## Project Structure

```
ledger/
  cmd/
    api/
      main.go
    worker/
      main.go
  internal/
    http/
      handlers.go
    repo/
      pg_repo.go
      mongo_repo.go
    queue/
      rabbit_publisher.go
      rabbit_consumer.go
      models.go
  migrations/
    init.sql
  docker-compose.yml
  Makefile
  README.md
  go.mod
  go.sum
```

---

## Table of Contents

- [cmd/api/main.go](#cmdapimaingo)
- [internal/repo/pg_repo.go](#internalrepopg_repogo)
- [internal/queue/rabbit_consumer.go](#internalqueuerabbit_consumergogo)
- [internal/http/handlers (not shown)](#internalhttphandlers-not-shown)
- [internal/repo/mongo_repo.go (not shown)](#internalrepomongo_repogo-not-shown)

---

## cmd/api/main.go

**Purpose:**  
Entry point for the Ledger API service.  
Initializes and wires together the HTTP server, PostgreSQL, MongoDB, and RabbitMQ.

**Key Responsibilities:**
- Reads configuration from environment variables.
- Connects to PostgreSQL (for account and transaction state).
- Connects to MongoDB (for ledger entries).
- Connects to RabbitMQ (for async transaction processing).
- Sets up HTTP routes using [go-chi/chi](https://github.com/go-chi/chi).
- Handles graceful shutdown on SIGINT/SIGTERM.

**How it works:**
- Sets up all connections and repositories.
- Mounts HTTP handlers.
- Starts the HTTP server on port 8080.
- Waits for termination signals and shuts down gracefully.

---

## internal/repo/pg_repo.go

**Purpose:**  
Implements the PostgreSQL-backed repository for managing accounts and transactions.

**Key Types & Methods:**

- `PGRepo struct`: Wraps a PostgreSQL connection pool.
- `CreateAccount`: Creates a new account with an initial balance.
- `GetAccount`: Retrieves the balance for a given account.
- `ApplyTransaction`: Applies a deposit or withdrawal, enforcing idempotency.
- `ApplyTransfer`: Transfers funds between accounts atomically and idempotently.

**Notes:**
- All write operations use transactions for atomicity.
- Idempotency is enforced via a `processed_messages` table and unique keys.
- Handles errors for insufficient funds, invalid types, and database issues.

---

## internal/queue/rabbit_consumer.go

**Purpose:**  
Consumes messages from RabbitMQ and applies ledger operations.

**Key Types & Methods:**

- `BalanceApplier` interface: Abstracts transaction and transfer application.
- `LedgerWriter` interface: Abstracts writing ledger entries.
- `Consumer struct`: Holds the RabbitMQ channel, queue name, and dependencies.
- `Start`: Main loop that consumes messages, applies transactions or transfers, writes to the ledger, and acknowledges or nacks messages.

**How it works:**
- Consumes messages from a RabbitMQ queue.
- Tries to unmarshal as a transaction or transfer.
- Applies the operation using the provided interfaces.
- Writes the result to the ledger.
- Handles message acknowledgment and requeueing on error.

---

## internal/http/handlers

**Purpose:**  
Defines HTTP handlers for the REST API endpoints.

**Responsibilities:**
- Routing and request validation.
- Invoking repository and queue operations.
- Formatting HTTP responses.

---

## internal/repo/mongo_repo.go

**Purpose:**  
Implements MongoDB-backed repository for storing ledger entries.

**Responsibilities:**
- Writing transaction and transfer records to MongoDB.
- Used for audit and reporting purposes.

---

## Dependencies

- [pgx](https://github.com/jackc/pgx) — PostgreSQL driver
- [mongo-driver](https://github.com/mongodb/mongo-go-driver) — MongoDB driver
- [amqp091-go](https://github.com/rabbitmq/amqp091-go) — RabbitMQ client
- [go-chi/chi](https://github.com/go-chi/chi) — HTTP router

---
