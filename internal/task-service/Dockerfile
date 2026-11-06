FROM golang:1.27-bookworm

WORKDIR /workspace

COPY internal/core/go.mod ./internal/core/go.mod
COPY internal/database-service/go.mod internal/database-service/go.sum ./internal/database-service/
COPY internal/task-service/go.mod internal/task-service/go.sum ./internal/task-service/

WORKDIR /workspace/internal/task-service
RUN go mod download

WORKDIR /workspace
COPY internal/core ./internal/core
COPY internal/database-service/api ./internal/database-service/api
COPY internal/task-service ./internal/task-service

WORKDIR /workspace/internal/task-service
RUN go build -o /app/task-service .

CMD ["/app/task-service"]
