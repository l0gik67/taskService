FROM golang:1.27-bookworm
LABEL authors="l0gik67"

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o /app/database-service .

CMD ["/app/database-service"]
