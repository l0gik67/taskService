package main

import (
	"net/http"
	"task-service/transport"
)

func main() {

	router := transport.NewRouter(&transport.TaskHandlerImpl{})

	http.ListenAndServe(":8080", router)

}
