package transport

import (
	"net/http"
	"task-service/service"

	"github.com/gorilla/mux"
)

func StartServer() {

}

func getRouter() *mux.Router {
	router := mux.NewRouter()

	return router
}

type TaskHandler struct {
	taskService *service.TaskService
}

func (taskHandler *TaskHandler) getAllTasks(w http.ResponseWriter, r *http.Request) {

}
