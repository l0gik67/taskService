FROM golang:1.27-bookworm
LABEL authors="l0gik67"

WORKDIR /workspace

COPY internal/core/go.mod ./internal/core/go.mod
COPY internal/database-service/go.mod internal/database-service/go.sum ./internal/database-service/

WORKDIR /workspace/internal/database-service
RUN go mod download

WORKDIR /workspace
COPY internal/core ./internal/core
COPY internal/database-service ./internal/database-service

WORKDIR /workspace/internal/database-service
RUN go build -o /app/database-service .

CMD ["/app/database-service"]
