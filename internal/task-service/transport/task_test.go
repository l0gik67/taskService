package transport

import (
	"context"
	"core/model"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"task-service/service"
	"testing"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type getByFiltersTaskServiceStub struct {
	service.TaskService
	called  bool
	ctx     context.Context
	filters model.TaskFilters
	tasks   []*model.Task
	err     error
}

type requestContextKey struct{}

type getByIDTaskServiceStub struct {
	service.TaskService
	called bool
	task   *model.Task
	err    error
}

func (stub *getByIDTaskServiceStub) GetById(context.Context, int64) (*model.Task, error) {
	stub.called = true
	return stub.task, stub.err
}

func (stub *getByFiltersTaskServiceStub) GetByFilters(ctx context.Context, filters model.TaskFilters) ([]*model.Task, error) {
	stub.called = true
	stub.ctx = ctx
	stub.filters = filters
	return stub.tasks, stub.err
}

func TestGetById_ReturnsNotFoundForGRPCNotFound(t *testing.T) {
	stub := &getByIDTaskServiceStub{err: status.Error(codes.NotFound, "task not found")}
	router := NewRouter(NewTaskHandler(stub))
	request := httptest.NewRequest(http.MethodGet, "/tasks/17", nil)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusNotFound {
		t.Fatalf("status code = %d, want %d", response.Code, http.StatusNotFound)
	}
	if !stub.called {
		t.Fatal("service was not called")
	}
}

func TestGetById_ReturnsBadRequestForInvalidIDWithoutCallingService(t *testing.T) {
	stub := &getByIDTaskServiceStub{}
	router := NewRouter(NewTaskHandler(stub))
	request := httptest.NewRequest(http.MethodGet, "/tasks/0", nil)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status code = %d, want %d", response.Code, http.StatusBadRequest)
	}
	if stub.called {
		t.Fatal("service must not be called for an invalid id")
	}
}

func TestGetById_ReturnsBadGatewayForUnavailableGRPCService(t *testing.T) {
	stub := &getByIDTaskServiceStub{err: status.Error(codes.Unavailable, "database-service unavailable")}
	router := NewRouter(NewTaskHandler(stub))
	request := httptest.NewRequest(http.MethodGet, "/tasks/17", nil)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusBadGateway {
		t.Fatalf("status code = %d, want %d", response.Code, http.StatusBadGateway)
	}
}

func TestGetById_ReturnsSnakeCaseJSON(t *testing.T) {
	stub := &getByIDTaskServiceStub{task: &model.Task{
		Id:        17,
		Title:     "Task",
		UserId:    42,
		Duration:  time.Minute,
		StartTime: time.Date(2026, time.November, 7, 12, 0, 0, 0, time.UTC),
		Priority:  3,
	}}
	router := NewRouter(NewTaskHandler(stub))
	request := httptest.NewRequest(http.MethodGet, "/tasks/17", nil)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status code = %d, want %d", response.Code, http.StatusOK)
	}
	var body map[string]json.RawMessage
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	for _, key := range []string{"id", "title", "user_id", "duration", "start_time", "priority"} {
		if _, ok := body[key]; !ok {
			t.Fatalf("response is missing %q: %v", key, body)
		}
	}
	for _, key := range []string{"Id", "UserId", "StartTime"} {
		if _, ok := body[key]; ok {
			t.Fatalf("response contains Go field name %q: %v", key, body)
		}
	}
}

func TestGetByFilters_MissingUserIDReturnsBadRequestWithoutCallingService(t *testing.T) {
	stub := &getByFiltersTaskServiceStub{}
	router := NewRouter(&TaskHandlerImpl{taskService: stub})
	request := httptest.NewRequest(http.MethodGet, "/tasks", nil)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status code = %d, want %d", response.Code, http.StatusBadRequest)
	}
	if !strings.Contains(response.Body.String(), model.ErrUserIDFilterRequired.Error()) {
		t.Fatalf("response body = %q, want missing user_id message", response.Body.String())
	}
	if stub.called {
		t.Fatal("service must not be called without the required user_id filter")
	}
}

func TestGetByFilters_ValidUserIDCallsServiceAndReturnsTasks(t *testing.T) {
	want := []*model.Task{
		{Id: 7, UserId: 42, Title: "First task", Priority: 1},
		{Id: 8, UserId: 42, Title: "Second task", Priority: 2},
	}
	stub := &getByFiltersTaskServiceStub{tasks: want}
	router := NewRouter(&TaskHandlerImpl{taskService: stub})
	request := httptest.NewRequest(http.MethodGet, "/tasks?user_id=42", nil)
	request = request.WithContext(context.WithValue(request.Context(), requestContextKey{}, "request context"))
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status code = %d, want %d", response.Code, http.StatusOK)
	}
	if !stub.called {
		t.Fatal("service was not called")
	}
	if stub.filters.UserID == nil || *stub.filters.UserID != 42 {
		t.Fatalf("service filters = %+v, want user_id 42", stub.filters)
	}
	if stub.ctx.Value(requestContextKey{}) != "request context" {
		t.Fatal("request context value was not passed to the service")
	}
	if contentType := response.Header().Get("Content-Type"); contentType != "application/json" {
		t.Fatalf("Content-Type = %q, want application/json", contentType)
	}

	var got []*model.Task
	if err := json.NewDecoder(response.Body).Decode(&got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(got) != len(want) {
		t.Fatalf("response task count = %d, want %d", len(got), len(want))
	}
	for index := range want {
		if *got[index] != *want[index] {
			t.Fatalf("response task[%d] = %+v, want %+v", index, got[index], want[index])
		}
	}
}

func TestGetByFilters_InvalidUserIDReturnsBadRequestWithoutCallingService(t *testing.T) {
	for _, testCase := range []struct {
		name        string
		query       string
		wantMessage string
	}{
		{
			name:        "not an integer",
			query:       "user_id=invalid",
			wantMessage: `query parameter "user_id" must be a valid integer`,
		},
		{
			name:        "zero",
			query:       "user_id=0",
			wantMessage: model.ErrUserIDFilterInvalid.Error(),
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			stub := &getByFiltersTaskServiceStub{}
			router := NewRouter(&TaskHandlerImpl{taskService: stub})
			request := httptest.NewRequest(http.MethodGet, "/tasks?"+testCase.query, nil)
			response := httptest.NewRecorder()

			router.ServeHTTP(response, request)

			if response.Code != http.StatusBadRequest {
				t.Fatalf("status code = %d, want %d", response.Code, http.StatusBadRequest)
			}
			if !strings.Contains(response.Body.String(), testCase.wantMessage) {
				t.Fatalf("response body = %q, want %q", response.Body.String(), testCase.wantMessage)
			}
			if stub.called {
				t.Fatal("service must not be called for an invalid user_id filter")
			}
		})
	}
}
