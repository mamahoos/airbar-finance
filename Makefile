MIGRATIONS_DIR := internal/infrastructure/postgres/migrations
DATABASE_URL ?= postgres://airbar:airbar@localhost:5434/airbar_finance?sslmode=disable
PROTO_DIR := proto
GEN_DIR := internal/gen/financev1

TEST_DATABASE_URL ?= $(DATABASE_URL)

COMPOSE := docker compose -f docker-compose.yml
COMPOSE_DEV := $(COMPOSE) -f docker-compose.dev.yml
COMPOSE_STAGING := $(COMPOSE) -f docker-compose.staging.yml
COMPOSE_PROD := $(COMPOSE) -f docker-compose.prod.yml

.PHONY: up up-dev up-staging up-prod down migrate-up migrate-down migrate-status proto build test test-integration vet verify

up: ## Start dev postgres + redis only
	$(COMPOSE_DEV) up -d postgres-finance redis

up-dev: ## Start full dev stack (build app image)
	$(COMPOSE_DEV) up -d --build

up-staging: ## Deploy staging stack (requires IMAGE_TAG)
	@test -n "$(IMAGE_TAG)" || (echo "IMAGE_TAG is required" && exit 1)
	docker network create airbar-staging 2>/dev/null || true
	COMPOSE_PROJECT_NAME=airbar-finance-staging IMAGE_TAG=$(IMAGE_TAG) $(COMPOSE_STAGING) up -d --remove-orphans

up-prod: ## Deploy production stack (requires IMAGE_TAG)
	@test -n "$(IMAGE_TAG)" || (echo "IMAGE_TAG is required" && exit 1)
	docker network create airbar-prod 2>/dev/null || true
	COMPOSE_PROJECT_NAME=airbar-finance-prod IMAGE_TAG=$(IMAGE_TAG) $(COMPOSE_PROD) up -d --remove-orphans

down: ## Stop dev stack
	$(COMPOSE_DEV) down

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
