package main

import (
	"context"
	taskv1 "database-service/api/task/v1"
	"database-service/repository"
	"database-service/service"
	"database-service/transport"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	pool, err := newPool(context.Background())
	if err != nil {
		return err
	}
	defer pool.Close()

	repo := repository.NewTaskRepository(pool)
	taskService := service.NewTaskServiceImpl(repo)
	handler := transport.NewTaskHandler(taskService)

	serverAddress := os.Getenv("ADDRESS")
	listener, err := net.Listen("tcp", serverAddress)
	if err != nil {
		return fmt.Errorf("listen: %w", err)
	}
	defer listener.Close()

	grpcServer := grpc.NewServer()
	defer grpcServer.Stop()

	taskv1.RegisterTaskServiceServer(grpcServer, handler)

	log.Printf("gRPC listening on %s", listener.Addr())
	return grpcServer.Serve(listener)
}

func newPool(ctx context.Context) (*pgxpool.Pool, error) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		return nil, fmt.Errorf("DATABASE_URL is not set")
	}

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	return pool, nil
}
