MIGRATIONS_DIR := internal/infrastructure/postgres/migrations
DATABASE_URL ?= postgres://airbar:airbar@localhost:5434/airbar_finance?sslmode=disable

.PHONY: up down migrate-up migrate-down migrate-status

up:
	docker compose up -d postgres-finance redis

down:
	docker compose down

migrate-up:
	goose -dir $(MIGRATIONS_DIR) postgres "$(DATABASE_URL)" up

migrate-down:
	goose -dir $(MIGRATIONS_DIR) postgres "$(DATABASE_URL)" down

migrate-status:
	goose -dir $(MIGRATIONS_DIR) postgres "$(DATABASE_URL)" status
