package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

var storage Storage
var err error

func main() {
	storage, err = InitStorage()
	if err != nil {
		panic(err)
	}
	log.Println("Init server")
	router := mux.NewRouter()
	router.Use(LoggerMiddleware)

	router.HandleFunc("/", HomeHandler).Methods(http.MethodGet)

	var ol OperationLogin
	router.Handle("/users/login", ol).Methods(http.MethodPost)

	server := &http.Server{
		Handler:      router,
		Addr:         "127.0.0.1:8000",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	log.Println("Server working")
	log.Fatal(server.ListenAndServe())
}
