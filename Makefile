include ./.env
export
export PATH := $(PATH):$(shell go env GOPATH)/bin
DATABASE_SERVICE_ADDRESS ?= localhost:$(DATABASE_SERVICE_PORT)
export DATABASE_SERVICE_ADDRESS

build-database-service:
	go -C internal/database-service build -o app .
run-database-service:
	./internal/database-service/app

build-task-service:
	go -C internal/task-service build -o app .
run-task-service:
	./internal/task-service/app

MIGRATION_DATABASE_URL := $(subst @$(DB_HOST):,@$(MIGRATION_DB_HOST):,$(DATABASE_URL))
migrate-up:
	@migrate -path "./internal/database-service/migrations" -database "$(MIGRATION_DATABASE_URL)" up

generate-grpc-code-database-service:
	protoc \
      -I . \
      -I /opt/homebrew/include \
      --go_out=. \
      --go_opt=paths=source_relative \
      --go-grpc_out=. \
      --go-grpc_opt=paths=source_relative \
      internal/database-service/api/task/v1/task.proto

deploy-service:
	docker compose --env-file .env -f docker-compose.yaml up -d --build --wait
	$(MAKE) migrate-up

PHONY: migrate-up
