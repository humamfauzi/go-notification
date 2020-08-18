package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
)

type LoginReply struct {
	Success bool   `json:"success"`
	Token   string `json:"token"`
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
