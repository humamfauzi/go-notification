package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

func main() {
	log.Println("Init server")
	router := mux.NewRouter()
	router.Use(LoggerMiddleware)
	router.HandleFunc("/", HomeHandler).Methods(http.MethodGet)
	router.HandleFunc("/users/login", LoginHandler).Methods(http.MethodPost)
	server := &http.Server{
		Handler:      router,
		Addr:         "127.0.0.1:8000",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	log.Println("Server working")
	log.Fatal(server.ListenAndServe())
}

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "HomePage")
	return
}

func LoggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("ACCESSED", r.Header.Get("User-Agent"))
		next.ServeHTTP(w, r)
	})
}

type UserLogin struct {
	Username string
	Password string
	Token    *string
}

func (ul UserLogin) AuthenticateLogin() bool {
	return true
}

func (ul *UserLogin) GenerateToken() {
	newToken := "newToken"
	ul.Token = &newToken
}

type LoginReply struct {
	Success bool   `json:"success"`
	Token   string `json:"token"`
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	body, _ := ioutil.ReadAll(r.Body)
	var loginProfile UserLogin
	json.Unmarshal(body, &loginProfile)

	if ok := loginProfile.AuthenticateLogin(); !ok {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	loginProfile.GenerateToken()

	loginReply := LoginReply{
		Success: true,
		Token:   *loginProfile.Token,
	}

	jsonReply, _ := json.Marshal(loginReply)
	w.WriteHeader(http.StatusOK)
	w.Write(jsonReply)
}
