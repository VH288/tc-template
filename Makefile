.PHONY: help run run-api migrate-up migrate-down migrate-status test test-unit tidy build lint

APP_NAME ?= tc-template
CONFIG ?= configs/config.yaml
GO ?= go

help:
	@echo "Targets:"
	@echo "  run / run-api       - run HTTP API (cmd/api)"
	@echo "  migrate-up          - run database migrations up"
	@echo "  migrate-down        - roll back database migrations"
	@echo "  migrate-status      - check migration status"
	@echo "  test                - run all tests"
	@echo "  test-unit           - run unit tests"
	@echo "  tidy                - run go mod tidy"
	@echo "  build               - build api and migrate binaries"
	@echo "  lint                - run go vet"

run: run-api

run-api:
	$(GO) run ./cmd/api

migrate-up:
	$(GO) run ./cmd/migrate -dir migrations up

migrate-down:
	$(GO) run ./cmd/migrate -dir migrations down

migrate-status:
	$(GO) run ./cmd/migrate -dir migrations status

test:
	$(GO) test ./... -count=1

test-unit:
	$(GO) test ./internal/domain/... ./internal/pkg/... -count=1 -short

tidy:
	$(GO) mod tidy

build:
	$(GO) build -o bin/api ./cmd/api
	$(GO) build -o bin/migrate ./cmd/migrate

lint:
	$(GO) vet ./...
