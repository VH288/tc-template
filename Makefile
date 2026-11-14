.PHONY: help run run-api run-grpc proto migrate-up migrate-down migrate-status test test-unit tidy build lint doc-sync docker-build docker-run

APP_NAME ?= tc-template
CONFIG ?= configs/config.yaml
GO ?= go
PROTOC ?= protoc
DOCKER ?= docker
IMAGE ?= $(APP_NAME):latest
HTTP_PORT ?= 8081
GRPC_PORT ?= 7000
DOCKER_NETWORK ?= tools_backend
DB_HOST ?= postgres

help:
	@echo "Targets:"
	@echo "  run / run-api       - run HTTP API (cmd/api)"
	@echo "  run-grpc            - run gRPC API (cmd/grpc)"
	@echo "  proto               - generate protobuf code"
	@echo "  migrate-up          - run database migrations up"
	@echo "  migrate-down        - roll back database migrations"
	@echo "  migrate-status      - check migration status"
	@echo "  test                - run all tests"
	@echo "  test-unit           - run unit tests"
	@echo "  tidy                - run go mod tidy"
	@echo "  build               - build api, grpc, and migrate binaries"
	@echo "  lint                - run go vet"
	@echo "  doc-sync            - synchronize OpenAPI specs into swagger catalog"
	@echo "  docker-build        - build container image ($(IMAGE))"
	@echo "  docker-run          - run image on $(DOCKER_NETWORK) with $(CONFIG)"
run: run-api

run-api:
	$(GO) run ./cmd/api

run-grpc:
	$(GO) run ./cmd/grpc

proto:
	$(PROTOC) -I internal/proto \
		--go_out=internal/proto --go_opt=module=github.com/VH288/tc-template/internal/proto \
		--go-grpc_out=internal/proto --go-grpc_opt=module=github.com/VH288/tc-template/internal/proto \
		internal/proto/sample/v1/sample.proto

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
	$(GO) build -o bin/grpc ./cmd/grpc
	$(GO) build -o bin/migrate ./cmd/migrate

lint:
	$(GO) vet ./...

doc-sync:
	$(GO) run scripts/sync-docs.go

docker-build:
	$(DOCKER) build -t $(IMAGE) .

docker-run:
	$(DOCKER) run --rm --network $(DOCKER_NETWORK) -p $(HTTP_PORT):8081 -p $(GRPC_PORT):7000 \
		--mount "type=bind,src=$(abspath $(CONFIG)),dst=/app/configs/config.yaml,readonly" \
		-e DB_HOST=$(DB_HOST) \
		$(IMAGE)
