include ./.env
export
export PATH := $(PATH):$(shell go env GOPATH)/bin

build-database-service:
	go -C internal/database-service build -o app .
run-database-service:
	./internal/database-service/app

migrate-up: DB_HOST := $(MIGRATION_DB_HOST)
migrate-up:
	migrate -path "./internal/database-service/migrations" -database "$$DATABASE_URL" up

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
