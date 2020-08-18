package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
)

type LoginRequest struct {
	UserName    string `json:"username"`
	Password    string `json:"password"`
	PhoneNumber string `json:"phone_number"`
}

type LoginReply struct {
	Code    int    `json:"code"`
	Success bool   `json:"success"`
	Token   string `json:"token"`
}

type OperationLogin struct{}

func (ol OperationLogin) SearchByUsername(username string) (UserProfile, error) {
	user, err := storage.Get(username)
	if err != nil {
		return UserProfile{}, err
	}
	return user, nil
}

func (ol OperationLogin) InterpretRequest(payload io.Reader) LoginRequest {
	body, _ := ioutil.ReadAll(payload)
	var loginRequest = LoginRequest{}
	json.Unmarshal(body, &loginRequest)
	return loginRequest
}

func (ol OperationLogin) AuthenticateLogin(lr LoginRequest) bool {
	userProfile, err := ol.SearchByUsername(lr.UserName)
	if err != nil {
		return false
	}
	if ok := userProfile.MatchUsernameAndPassword(lr.UserName, lr.Password); !ok {
		return false
	}
	return true

}

func (ol OperationLogin) GetToken(lr LoginRequest) (string, error) {
	userProfile, err := storage.Get(lr.UserName)
	if err != nil {
		return "", errors.New("ERR_TOKEN_NOT_FOUND")
	}
	return userProfile.UserCredential.Token, nil
}

func (ol OperationLogin) ComposeReply(w http.ResponseWriter, lr LoginReply) {
	w.WriteHeader(lr.Code)
	reply, _ := json.Marshal(lr)
	w.Write(reply)
	return
}

func (ol OperationLogin) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	loginRequest := ol.InterpretRequest(r.Body)
	if ok := ol.AuthenticateLogin(loginRequest); !ok {
		ol.ComposeReply(w, LoginReply{
			Code:    http.StatusForbidden,
			Success: false,
			Token:   "",
		})
		return
	}

	token, err := ol.GetToken(loginRequest)
	if err != nil {
		ol.ComposeReply(w, LoginReply{
			Code:    http.StatusForbidden,
			Success: false,
			Token:   "",
		})
		return
	}
	if err != nil {
		ol.ComposeReply(w, LoginReply{
			Code:    http.StatusInternalServerError,
			Success: false,
			Token:   "",
		})
		return
	}
	ol.ComposeReply(w, LoginReply{
		Code:    http.StatusOK,
		Success: true,
		Token:   token,
	})
	return
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
