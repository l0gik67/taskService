# taskService — Checklist application

`taskService` is an educational Go project for a task-management application similar in purpose to Microsoft To Do. The current goal is to build two cooperating microservices while keeping database access behind a dedicated service.

> The project is still in development. `database-service` is the most complete component, while the public API service, `task-service`, is currently only a code skeleton.

## Target architecture

```text
User
  │ HTTP
  ▼
task-service (public API)
  │ gRPC
  ▼
database-service
  │ SQL through pgx
  ▼
PostgreSQL
```

- `task-service` accepts public HTTP requests and forwards task operations to `database-service`.
- `database-service` owns task persistence and is the only application service that communicates with PostgreSQL.
- PostgreSQL stores the tasks.

The original requirements allowed HTTP or gRPC between the application services. The repository has already selected **gRPC**: its current contract is defined in [`internal/database-service/api/task/v1/task.proto`](internal/database-service/api/task/v1/task.proto).

## Current implementation status

### `database-service`: mostly implemented for the current model

The service currently includes:

- a runnable gRPC server;
- a PostgreSQL connection through `pgxpool`;
- `transport` → `service` → `repository` layers;
- create, get-by-ID, list, update, and delete operations;
- task-list filtering by a required positive `user_id`;
- an SQL migration for the `tasks` table;
- a protobuf contract and generated Go code;
- a Dockerfile and Docker Compose configuration alongside PostgreSQL;
- unit tests for part of the service and gRPC transport layers.

The task model currently implemented in Go, protobuf, and SQL is:

| Field | Purpose |
|---|---|
| `id` | unique task identifier |
| `title` | task title |
| `user_id` | user identifier |
| `duration_seconds` | duration in seconds |
| `start_time` | task start time |
| `priority` | priority from 1 to 100 |

### `task-service`: in progress

`task-service` is not runnable yet. It only contains an empty service type, a router skeleton, and an unfinished list handler. It still needs:

- a `main` entry point;
- its own Go module and completed dependency setup;
- implementations of the public HTTP routes;
- a gRPC client for `database-service`;
- request and response serialization;
- error handling and tests;
- a Dockerfile and an entry in `docker-compose.yaml`.

### Differences from the original requirements

The original assignment requires a task title, description, and completion state. The current database schema and protobuf model contain `user_id`, duration, start time, and priority, but do not contain a description or completion field.

Therefore, `database-service` is close to a working CRUD service for the **current** model, but it does not yet fully implement the assignment's domain model. In particular, the current gRPC `UpdateTask` method updates a complete task and is not the required “mark as done” operation because no `done` or `completed` field exists yet.

## Public HTTP API

The assignment defines the following public contract. It belongs in `task-service` and is **not implemented yet**:

| Method | Route | Purpose |
|---|---|---|
| `POST` | `/create` | create a task from a title and description |
| `GET` | `/list` | return the task list |
| `DELETE` | `/delete` | delete a task; its ID is passed in the request body |
| `PUT` | `/done` | mark a task as completed; its ID is passed in the request body |

The assignment does not define exact JSON bodies, the ID type, validation rules, or HTTP status codes. These details still need to be designed. The required ID placement for `/delete` and `/done` should remain in the request body.

## Internal gRPC API

The current `database-service` contract provides:

- `CreateTask`;
- `UpdateTask`;
- `DeleteTask`;
- `GetTaskById`;
- `ListTasks`, which requires a positive `user_id`.

This contract reflects the model that exists in the repository today. It will need to evolve when task descriptions and completion state are added.

## Technology stack

- Go `1.27.3`, as declared by the module and Docker base image;
- gRPC and Protocol Buffers;
- PostgreSQL `18-bookworm`;
- `pgx/v5` and `pgxpool`;
- Docker and Docker Compose;
- SQL migrations executed with the `golang-migrate` CLI;
- Make targets for building, running, migrating, and generating gRPC code.

The `task-service` skeleton imports `gorilla/mux`, but that service is not yet configured as a buildable Go module.

## Repository structure

```text
.
├── Makefile
├── docker-compose.yaml
└── internal
    ├── database-service
    │   ├── api/task/v1       # protobuf contract and generated code
    │   ├── migrations        # PostgreSQL migrations
    │   ├── model             # internal task model and list filter
    │   ├── repository        # SQL and PostgreSQL access
    │   ├── service           # business logic
    │   ├── service-test      # service-layer tests
    │   ├── transport         # gRPC server implementation and tests
    │   ├── Dockerfile
    │   └── main.go
    └── task-service
        ├── service           # initial public API service skeleton
        └── transport         # initial HTTP router skeleton
```

## Configuration

The `Makefile` loads the root `.env` file. This file is excluded from Git and must be created locally. The repository currently refers to these variable names:

- `DB_USER`, `DB_PASSWORD`, `DB_NAME`, `DB_HOST`, and `DB_PORT`;
- `MIGRATION_DB_HOST`;
- `DATABASE_SERVICE_PORT` and `ADDRESS`;
- `DATABASE_URL`.

Their values must be consistent across PostgreSQL, `database-service`, port mapping, and the migration CLI. Do not commit a real `.env` file or database credentials.

## Running the current project

### Docker Compose

The current deployment target requires Docker, Docker Compose, `make`, the `golang-migrate` CLI, and a populated `.env` file:

```bash
make deploy-service
```

This Make target builds and starts PostgreSQL and `database-service`, waits for the PostgreSQL health check, and then applies the migration. It does not start a public HTTP API because `task-service` has not been added to Compose yet.

### Local `database-service` build

Local execution additionally requires a Go version compatible with `go.mod` and a reachable PostgreSQL instance:

```bash
make build-database-service
make migrate-up
make run-database-service
```

`make run-database-service` runs the previously built `internal/database-service/app` binary and uses configuration loaded from `.env`.

### Tests

The database-service tests belong to its standalone Go module:

```bash
go -C internal/database-service test ./...
```

### Protobuf code generation

After changing the `.proto` contract, use the repository target:

```bash
make generate-grpc-code-database-service
```

The current Makefile expects `protoc`, the `protoc-gen-go` and `protoc-gen-go-grpc` plugins, and system protobuf definitions under `/opt/homebrew/include`.

## Next steps

1. **Align `database-service` with the assignment.** Confirm the final task model, add description and completion fields, add the required migration and completion operation, and extend validation and test coverage.
2. **Implement `task-service`.** Add a runnable HTTP server, the four required routes, a gRPC client, HTTP-to-gRPC mapping, error handling, tests, and containerization.
3. **Run the complete two-service flow.** Start both application services with PostgreSQL and verify create, list, delete, and mark-as-done operations through the public HTTP API.

This README intentionally limits the project scope to these two application services and the PostgreSQL dependency.
