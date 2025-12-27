include .env

LOCAL_BIN := $(ROOT_DIR)/bin

export ROOT_DIR := $(CURDIR)
export GOOSE_MIGRATION_DIR := $(ROOT_DIR)/internal/migrations

export GOOSE_MIGRATION_DIR
export GOOSE_DRIVER
export GOOSE_DBSTRING

install-golangci-lint:
	mkdir -p ./bin
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b ./bin

lint:
	$(LOCAL_BIN)/golangci-lint run ./... --config .golangci.pipeline.yaml

install-deps:
	GOBIN=$(LOCAL_BIN) go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	GOBIN=$(LOCAL_BIN) go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

get-deps:
	go get -u google.golang.org/protobuf/cmd/protoc-gen-go
	go get -u google.golang.org/grpc/cmd/protoc-gen-go-grpc

generate:
	make generate-auth-api

generate-auth-api:
	mkdir -p pkg/auth_v1
	protoc --proto_path api/auth_v1 \
	--go_out=pkg/auth_v1 --go_opt=paths=source_relative \
	--plugin=protoc-gen-go=bin/protoc-gen-go \
	--go-grpc_out=pkg/auth_v1 --go-grpc_opt=paths=source_relative \
	--plugin=protoc-gen-go-grpc=bin/protoc-gen-go-grpc \
	api/auth_v1/auth.proto

docker/up:
	docker compose up -d

docker/down:
	docker compose down

migrate/create:
	chmod +x ./scripts/migrations/create.sh && \
	sh ./scripts/migrations/create.sh

migrate/up:
	goose up --dir $(GOOSE_MIGRATION_DIR)

migrate/down:
	goose down --dir $(GOOSE_MIGRATION_DIR)

server/run:
	go run ./cmd/server/main.go

client/run:
	go run ./cmd/client/main.go

sqlc/gen:
	sqlc generate -f ./internal/sqlc.yaml