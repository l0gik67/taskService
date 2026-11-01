package transport

import (
	"net/http"

	"github.com/gorilla/mux"
)

func NewRouter(handler TaskHandler) *mux.Router {
	router := mux.NewRouter()

	router.HandleFunc("/tasks/{id}", handler.GetById).Methods(http.MethodGet)
	router.HandleFunc("/tasks", handler.GetByUserId).Methods(http.MethodGet)
	router.HandleFunc("/tasks", handler.Create).Methods(http.MethodPost)
	router.HandleFunc("/tasks", handler.Update).Methods(http.MethodPut)
	router.HandleFunc("/tasks", handler.Delete).Methods(http.MethodDelete)

	return router
}

type TaskHandler interface {
	GetById(w http.ResponseWriter, r *http.Request)
	GetByUserId(w http.ResponseWriter, r *http.Request)
	Create(w http.ResponseWriter, r *http.Request)
	Update(w http.ResponseWriter, r *http.Request)
	Delete(w http.ResponseWriter, r *http.Request)
}
