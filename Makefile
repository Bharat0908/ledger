build:
	docker compose build

up:
	docker compose up -d --build

down:
	docker compose down -v

migrate:
	docker compose exec -T postgres psql -U postgres -d ledger -f /docker-entrypoint-initdb.d/schema.sql

smoke:
	bash smoke_test.sh

lint:
	golangci-lint run

test:
	go test ./...
