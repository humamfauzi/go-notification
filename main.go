package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
)

type Storage interface {
	Get(string) (UserLogin, error)
}

type Environement string

func (env Environement) getCurrentEnv() string {
	return os.Getenv("GO_ENV")
}
func (env Environement) IsTest() bool {
	return env.getCurrentEnv() == "test"
}

type InMemory struct {
	Members []UserLogin
}

func (im InMemory) Get(userName string) (UserLogin, error) {
	for _, user := range im.Members {
		if user.Username == userName {
			return user, nil
		}
	}
	return UserLogin{}, errors.New("USER_NOT_FOUND")
}

func CreateInMemoryStorage() Storage {
	token1 := "HEHE123"
	token2 := "1253NNMK"
	newStorage := InMemory{
		Members: []UserLogin{
			UserLogin{
				Username: "Hello123",
				Password: "AXZ098",
				Token:    &token1,
			},
			UserLogin{
				Username: "Bye456",
				Password: "IOP678",
				Token:    &token2,
			},
		},
	}
	return newStorage
}
func initStorage() (Storage, error) {
	var env Environement
	if env.IsTest() {
		return CreateInMemoryStorage(), nil
	} else {
		return InMemory{}, errors.New("unsupported env")
	}
}

var storage Storage
var err error

func main() {
	storage, err = initStorage()
	if err != nil {
		panic(err)
	}
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
	Username string `json:"username"`
	Password string `json:"password"`
	Token    *string
}

func (ul *UserLogin) InterpretPayload(payload io.Reader) {
	body, _ := ioutil.ReadAll(payload)
	json.Unmarshal(body, ul)
}

func (ul UserLogin) AuthenticateLogin() bool {
	_, err := storage.Get(ul.Username)
	if err != nil {
		return false
	}
	return true
}

func (ul *UserLogin) GenerateToken() {
	user, err := storage.Get(ul.Username)
	if err != nil {
		ul.Token = nil
	}
	ul.Token = user.Token
}

type LoginReply struct {
	Success bool   `json:"success"`
	Token   string `json:"token"`
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	loginProfile := UserLogin{}
	loginProfile.InterpretPayload(r.Body)
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
