# Ledger API Service

This is the main entry point for the Ledger API service. It initializes and connects the core infrastructure components, sets up HTTP routes, and starts the API server.

---

## Features

- **PostgreSQL Integration:** Uses a connection pool for account and transaction storage.
- **MongoDB Integration:** Stores ledger entries for audit/history.
- **RabbitMQ Integration:** Publishes transaction messages for asynchronous processing.
- **HTTP API:** Exposes endpoints for account and transaction operations.
- **Graceful Shutdown:** Handles SIGINT/SIGTERM for clean server shutdown.

---

## Environment Variables

- `POSTGRES_DSN` — PostgreSQL connection string (default: `postgres://postgres:postgres@postgres:5432/ledger`)
- `MONGO_URI` — MongoDB connection string (default: `mongodb://mongo:27017`)
- `RABBITMQ_URL` — RabbitMQ connection string (default: `amqp://guest:guest@rabbitmq:5672/`)

---

## Main Components

- **PostgreSQL:** Used for account and transaction state (via `repo.PGRepo`).
- **MongoDB:** Used for storing ledger entries (via `repo.MongoRepo`).
- **RabbitMQ:** Used for publishing transaction events (via `queue.Publisher`).
- **HTTP Handlers:** Provided by `internal/http/handlers`, mounted using [chi router](https://github.com/go-chi/chi).

---

## Startup Sequence

1. **Connect to PostgreSQL** using `pgxpool`.
2. **Connect to MongoDB** using the official MongoDB driver.
3. **Connect to RabbitMQ** and declare the exchange and queue.
4. **Initialize repositories** for Postgres and MongoDB.
5. **Create the HTTP handler** and mount routes.
6. **Start the HTTP server** on port `8080`.
7. **Wait for shutdown signal** and gracefully stop the server.

---

## Example Usage

Start the service (with Docker Compose