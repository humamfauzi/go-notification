package main

import "log"

type UserProfile struct {
	UserDetail
	UserCredential
}

type UserDetail struct {
	UserId      string
	Username    string
	FirstName   string
	LastName    string
	PhoneNumber string
	Email       string
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
