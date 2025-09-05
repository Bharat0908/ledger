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
