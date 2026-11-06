package transport_test

import (
	"context"
	taskv1 "database-service/api/task/v1"
	"database-service/model"
	"database-service/repository"
	"database-service/service"
	"database-service/transport"
	"errors"
	"net"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type listRepository struct {
	repository.TaskRepository
	calls  int
	filter model.TaskFilter
	tasks  []*model.Task
	err    error
}

func (repo *listRepository) List(filter model.TaskFilter) ([]*model.Task, error) {
	repo.calls++
	repo.filter = filter
	return repo.tasks, repo.err
}

func newListClient(t *testing.T, repo repository.TaskRepository) taskv1.TaskServiceClient {
	t.Helper()
	listener := bufconn.Listen(1024 * 1024)
	server := grpc.NewServer()
	taskv1.RegisterTaskServiceServer(server, transport.NewTaskHandler(service.NewTaskServiceImpl(repo)))
	serveErr := make(chan error, 1)
	go func() { serveErr <- server.Serve(listener) }()
	t.Cleanup(func() {
		server.Stop()
		listener.Close()
		if err := <-serveErr; err != nil && !errors.Is(err, grpc.ErrServerStopped) {
			t.Errorf("serve gRPC: %v", err)
		}
	})

	conn, err := grpc.NewClient("passthrough:///task-service-test",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
			return listener.DialContext(ctx)
		}),
	)
	if err != nil {
		t.Fatalf("create gRPC client: %v", err)
	}
	t.Cleanup(func() { conn.Close() })
	return taskv1.NewTaskServiceClient(conn)
}

func TestListTasks_RequiresPositiveUserID(t *testing.T) {
	for _, tc := range []struct {
		name   string
		userID int64
	}{
		{name: "missing userId", userID: 0},
		{name: "negative userId", userID: -1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			repo := &listRepository{}
			client := newListClient(t, repo)
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			_, err := client.ListTasks(ctx, &taskv1.ListTasksRequest{UserId: tc.userID})
			if status.Code(err) != codes.InvalidArgument {
				t.Fatalf("code = %v, want InvalidArgument; error: %v", status.Code(err), err)
			}
			if repo.calls != 0 {
				t.Fatal("invalid request must not reach the repository")
			}
		})
	}
}

func TestListTasks_RejectsNilRequest(t *testing.T) {
	handler := transport.NewTaskHandler(service.NewTaskServiceImpl(nil))
	_, err := handler.ListTasks(context.Background(), nil)
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("code = %v, want InvalidArgument", status.Code(err))
	}
}

func TestListTasks_PassesFilterAndReturnsTasks(t *testing.T) {
	start := time.Date(2026, time.November, 6, 12, 0, 0, 0, time.UTC)
	repo := &listRepository{tasks: []*model.Task{{
		Id: 17, UserId: 42, Title: "Task", Duration: 2 * time.Minute, StartTime: start, Priority: 3,
	}}}
	client := newListClient(t, repo)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	response, err := client.ListTasks(ctx, &taskv1.ListTasksRequest{UserId: 42})
	if err != nil {
		t.Fatalf("ListTasks: %v", err)
	}
	if repo.calls != 1 || repo.filter.UserID != 42 {
		t.Fatalf("repository calls=%d filter=%+v; want one call for user 42", repo.calls, repo.filter)
	}
	want := &taskv1.ListTasksResponse{Tasks: []*taskv1.Task{{
		Id: 17, UserId: 42, Title: "Task", DurationSeconds: 120, StartTime: timestamppb.New(start), Priority: 3,
	}}}
	if !proto.Equal(response, want) {
		t.Fatalf("response = %v, want %v", response, want)
	}
}

func TestListTasks_EmptyResultIsNotAnError(t *testing.T) {
	client := newListClient(t, &listRepository{})
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	response, err := client.ListTasks(ctx, &taskv1.ListTasksRequest{UserId: 42})
	if err != nil {
		t.Fatalf("ListTasks: %v", err)
	}
	if len(response.Tasks) != 0 {
		t.Fatalf("expected empty tasks, got %v", response.Tasks)
	}
}

func TestListTasks_RepositoryErrorReturnsInternal(t *testing.T) {
	client := newListClient(t, &listRepository{err: errors.New("database unavailable")})
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := client.ListTasks(ctx, &taskv1.ListTasksRequest{UserId: 42})
	if status.Code(err) != codes.Internal {
		t.Fatalf("code = %v, want Internal; error: %v", status.Code(err), err)
	}
}
