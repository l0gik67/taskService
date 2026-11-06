package transport

import (
	"core/dto"
	"core/model"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"task-service/service"

	"github.com/gorilla/mux"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func NewRouter(handler TaskHandler) *mux.Router {
	router := mux.NewRouter()

	router.HandleFunc("/tasks/{id}", handler.GetById).Methods(http.MethodGet)
	router.HandleFunc("/tasks", handler.GetByFilters).Methods(http.MethodGet)
	router.HandleFunc("/tasks", handler.Create).Methods(http.MethodPost)
	router.HandleFunc("/tasks", handler.Update).Methods(http.MethodPut)
	router.HandleFunc("/tasks/{id}", handler.Delete).Methods(http.MethodDelete)

	return router
}

type TaskHandler interface {
	GetById(w http.ResponseWriter, r *http.Request)
	GetByFilters(w http.ResponseWriter, r *http.Request)
	Create(w http.ResponseWriter, r *http.Request)
	Update(w http.ResponseWriter, r *http.Request)
	Delete(w http.ResponseWriter, r *http.Request)
}

type TaskHandlerImpl struct {
	taskService service.TaskService
}

func NewTaskHandler(taskService service.TaskService) *TaskHandlerImpl {
	return &TaskHandlerImpl{taskService: taskService}
}

func (handler *TaskHandlerImpl) GetById(w http.ResponseWriter, r *http.Request) {
	idString := mux.Vars(r)["id"]
	if idString == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("id is empty"))
		return
	}
	id, err := strconv.ParseInt(idString, 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}
	if id <= 0 {
		http.Error(w, "id must be greater than zero", http.StatusBadRequest)
		return
	}
	task, err := handler.taskService.GetById(r.Context(), id)
	if err != nil {
		writeGetByIDError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, task)
}

func writeGetByIDError(w http.ResponseWriter, err error) {
	if errors.Is(err, model.ErrTaskNotFound) || status.Code(err) == codes.NotFound {
		http.Error(w, model.ErrTaskNotFound.Error(), http.StatusNotFound)
		return
	}

	switch status.Code(err) {
	case codes.InvalidArgument:
		http.Error(w, "invalid request", http.StatusBadRequest)
	case codes.Unavailable:
		http.Error(w, "database service unavailable", http.StatusBadGateway)
	default:
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}
}

func (handler *TaskHandlerImpl) GetByFilters(w http.ResponseWriter, r *http.Request) {
	filters, err := taskFiltersFromRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	tasks, err := handler.taskService.GetByFilters(r.Context(), filters)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	writeJSON(w, http.StatusOK, tasks)
}

func (handler *TaskHandlerImpl) Create(w http.ResponseWriter, r *http.Request) {
	taskDto := &dto.TaskCreateDto{}
	err := json.NewDecoder(r.Body).Decode(taskDto)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}
	taskModel := &model.Task{
		UserId:    taskDto.UserId,
		Title:     taskDto.Title,
		StartTime: taskDto.StartTime,
		Duration:  taskDto.Duration,
		Priority:  taskDto.Priority,
	}
	task, err := handler.taskService.Create(r.Context(), taskModel)

	if err != nil {
		// TODO: Once service errors are defined, map validation errors to 400 and internal failures to 500.
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}
	writeJSON(w, http.StatusCreated, task)
}

func (handler *TaskHandlerImpl) Update(w http.ResponseWriter, r *http.Request) {
	taskDto := &dto.TaskUpdateDto{}
	err := json.NewDecoder(r.Body).Decode(taskDto)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}
	if taskDto.Id <= 0 {
		http.Error(w, "id must be greater than zero", http.StatusBadRequest)
		return
	}
	taskModel := &model.Task{
		Id:        taskDto.Id,
		UserId:    taskDto.UserId,
		Title:     taskDto.Title,
		StartTime: taskDto.StartTime,
		Duration:  taskDto.Duration,
		Priority:  taskDto.Priority,
	}
	err = handler.taskService.Update(r.Context(), taskModel)
	if err != nil {
		// TODO: Once service errors are defined, map validation errors to 400, not found to 404, and internal failures to 500.
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}
}

func (handler *TaskHandlerImpl) Delete(w http.ResponseWriter, r *http.Request) {
	idString := mux.Vars(r)["id"]
	if idString == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("id is empty"))
		return
	}
	id, err := strconv.ParseInt(idString, 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}
	if id <= 0 {
		http.Error(w, "id must be greater than zero", http.StatusBadRequest)
		return
	}
	err = handler.taskService.Delete(r.Context(), id)
	if err != nil {
		// TODO: Once service errors are defined, map not found to 404 and internal failures to 500.
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func writeJSON(w http.ResponseWriter, statusCode int, value any) {
	body, err := json.Marshal(value)
	if err != nil {
		log.Printf("encode JSON response: %v", err)
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if _, err := w.Write(body); err != nil {
		log.Printf("write JSON response: %v", err)
	}
}
