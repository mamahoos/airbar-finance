MIGRATIONS_DIR := internal/infrastructure/postgres/migrations
DATABASE_URL ?= postgres://airbar:airbar@localhost:5434/airbar_finance?sslmode=disable
PROTO_DIR := proto
GEN_DIR := internal/gen/financev1

TEST_DATABASE_URL ?= $(DATABASE_URL)

.PHONY: up down migrate-up migrate-down migrate-status proto build test test-integration vet verify

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

proto:
	@command -v protoc >/dev/null || (echo "protoc not found; install protobuf-compiler" && exit 1)
	@command -v protoc-gen-go >/dev/null || (echo "protoc-gen-go not found; go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.36.5" && exit 1)
	@command -v protoc-gen-go-grpc >/dev/null || (echo "protoc-gen-go-grpc not found; go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.5.1" && exit 1)
	mkdir -p $(GEN_DIR)
	protoc -I $(PROTO_DIR) \
		--go_out=$(GEN_DIR) --go_opt=paths=source_relative \
		--go-grpc_out=$(GEN_DIR) --go-grpc_opt=paths=source_relative \
		$(PROTO_DIR)/airbar_finance_v1.proto

build:
	go build -buildvcs=false -o bin/airbar-finance ./cmd/server

test:
	go test ./...

vet:
	go vet ./...

verify: vet test build

test-integration:
	TEST_DATABASE_URL="$(TEST_DATABASE_URL)" go test -tags=integration ./internal/infrastructure/postgres/repository/ -run Integration -v
