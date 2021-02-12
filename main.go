package main

import (
	"log"
	"net/http"
	// "time"

	"github.com/gorilla/mux"
	dba "github.com/humamfauzi/go-notification/database"
	"github.com/humamfauzi/go-notification/handler"
)

const (
	queryMapDir = "database/queryMap.json"
)

func main() {
	connProp := dba.MysqlDatabaseAccess{
		Username: "root",
		Password: "",
		Protocol: "tcp",
		Address: "localhost",
		DBName: "try1",
	}
	connDB, err := connProp.ConnectDatabase()
	if err != nil {
		panic(err)
	}
	defer connDB.Close()
	if err := dba.ConvertJsonToQueryMap(queryMapDir); err != nil {
		panic("Failed to read query map")
	}
	log.Println("OK")
	
	log.Println("Init server")
	router := mux.NewRouter()
	router.Use(LoggerMiddleware)

	router.HandleFunc("/user/create", CreateUserHandler).Methods(http.MethodPost)
	router.HandleFunc("/user/{id}", UpdateUserHandler).Methods(http.MethodPut)
	router.HandleFunc("/user/{id}", DeleteUserHandler).Methods(http.MethodDelete)

	router.HandleFunc("/topic/create", CreateTopicHandler).Methods(http.MethodPost)
	router.HandleFunc("/topic", GetTopicHandler).Methods(http.MethodsGet)
	router.HandleFunc("/subscribe/{id_topic}", CreateSubscribeHandler).Methods(http.MethodPost)

	router.HandleFunc("/notification/{id_topic}", CreateNotificationHandler).Methods(http.MethodPost)
	router.HandleFunc("/notification", GetNotificationHandler).Methods(http.MethodGet)


	server := &http.Server{
		Handler:      router,
		Addr:         "127.0.0.1:8000",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	log.Println("Server working")
	log.Fatal(server.ListenAndServe())
}
