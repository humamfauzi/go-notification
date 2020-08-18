package main

import "log"

type UserProfile struct {
	UserDetail
	UserCredential
}

type UserDetail struct {
	UserId      string `json:"user_id"`
	Username    string `json:"username"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	PhoneNumber string `json:"phone_number"`
	Email       string `json:"email`
}

type UserCredential struct {
	UserId   string
	Password string
	Token    string
}

func (up UserProfile) MatchUsernameAndPassword(userName string, password string) bool {
	stored := up.UserDetail.Username + ":" + up.UserCredential.Password
	requested := userName + ":" + password
	log.Println(stored, requested)
	return stored == requested
}

func (up UserProfile) GetToken() string {
	return up.UserCredential.Token
}
