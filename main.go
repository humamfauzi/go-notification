package main

import (
	"log"
	"net/http"
	"time"

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
	router.Use(handler.LoggerMiddleware)

	router.HandleFunc("/user", handler.CreateUserHandler).Methods(http.MethodPost)

	loginHandler := handler.LoginOps{}
	router.Handle("/user/login", loginHandler).Methods(http.MethodPost)

	checkLoginHandler := handler.CheckLogin{}
	router.Handle("/user/check", checkLoginHandler).Methods(http.MethodPost)

	router.HandleFunc("/user", handler.UpdateUserHandler).Methods(http.MethodPut)
	router.HandleFunc("/user", handler.DeleteUserHandler).Methods(http.MethodPut)

	router.HandleFunc("/topics", handler.CreateTopicHandler).Methods(http.MethodPost)
	router.HandleFunc("/topics", handler.GetTopicHandler).Methods(http.MethodGet)
	router.HandleFunc("/subscribe", handler.CreateSubscribeHandler).Methods(http.MethodPost)

	createNotificationHandler := handler.CreateNotification{}
	router.Handle("/notification", createNotificationHandler).Methods(http.MethodPost)
	router.HandleFunc("/notification", handler.GetNotificationHandler).Methods(http.MethodGet)


	server := &http.Server{
		Handler:      router,
		Addr:         "127.0.0.1:8000",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	log.Println("Server working")
	log.Fatal(server.ListenAndServe())
}
