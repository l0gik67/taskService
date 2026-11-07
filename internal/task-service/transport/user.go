package transport

import (
	"core/dto"
	"core/model"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"task-service/service"

	"github.com/gorilla/mux"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UserHandler struct {
	userService service.UserService
}

func NewUserHandler(userService service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

func RegisterUserRoutes(router *mux.Router, handler *UserHandler) {
	router.HandleFunc("/users", handler.Create).Methods(http.MethodPost)
	router.HandleFunc("/users", handler.GetAll).Methods(http.MethodGet)
	router.HandleFunc("/users/{id}", handler.GetById).Methods(http.MethodGet)
	router.HandleFunc("/users/{id}", handler.Update).Methods(http.MethodPut)
	router.HandleFunc("/users/{id}", handler.Delete).Methods(http.MethodDelete)
}

func (handler *UserHandler) Create(writer http.ResponseWriter, request *http.Request) {
	dtoValue := &dto.UserCreateDto{}
	if err := json.NewDecoder(request.Body).Decode(dtoValue); err != nil {
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}
	user, err := handler.userService.Create(request.Context(), &model.User{Name: dtoValue.Name})
	if err != nil {
		writeUserError(writer, err)
		return
	}
	writeJSON(writer, http.StatusCreated, user)
}

func (handler *UserHandler) GetById(writer http.ResponseWriter, request *http.Request) {
	id, err := positivePathId(request, "id")
	if err != nil {
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}
	user, err := handler.userService.GetById(request.Context(), id)
	if err != nil {
		writeUserError(writer, err)
		return
	}
	writeJSON(writer, http.StatusOK, user)
}

func (handler *UserHandler) GetAll(writer http.ResponseWriter, request *http.Request) {
	users, err := handler.userService.GetAll(request.Context())
	if err != nil {
		writeUserError(writer, err)
		return
	}
	writeJSON(writer, http.StatusOK, users)
}

func (handler *UserHandler) Update(writer http.ResponseWriter, request *http.Request) {
	id, err := positivePathId(request, "id")
	if err != nil {
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}
	dtoValue := &dto.UserUpdateDto{}
	if err := json.NewDecoder(request.Body).Decode(dtoValue); err != nil {
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}
	user, err := handler.userService.Update(request.Context(), &model.User{Id: id, Name: dtoValue.Name})
	if err != nil {
		writeUserError(writer, err)
		return
	}
	writeJSON(writer, http.StatusOK, user)
}

func (handler *UserHandler) Delete(writer http.ResponseWriter, request *http.Request) {
	id, err := positivePathId(request, "id")
	if err != nil {
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}
	if err := handler.userService.Delete(request.Context(), id); err != nil {
		writeUserError(writer, err)
		return
	}
	writer.WriteHeader(http.StatusNoContent)
}

func writeUserError(writer http.ResponseWriter, err error) {
	if errors.Is(err, model.ErrUserNotFound) {
		http.Error(writer, err.Error(), http.StatusNotFound)
		return
	}
	if errors.Is(err, model.ErrUserInUse) {
		http.Error(writer, err.Error(), http.StatusConflict)
		return
	}
	switch status.Code(err) {
	case codes.InvalidArgument, codes.Unknown:
		http.Error(writer, err.Error(), http.StatusBadRequest)
	case codes.NotFound:
		http.Error(writer, err.Error(), http.StatusNotFound)
	case codes.FailedPrecondition:
		http.Error(writer, err.Error(), http.StatusConflict)
	case codes.Unavailable:
		http.Error(writer, "database service unavailable", http.StatusBadGateway)
	default:
		log.Printf("user request failed: %v", err)
		http.Error(writer, "internal server error", http.StatusInternalServerError)
	}
}
