package main

import "errors"

type InMemory struct {
	Members []LoginRequest
}

func (im InMemory) Get(userName string) (LoginRequest, error) {
	for _, user := range im.Members {
		if user.UserName == userName {
			return user, nil
		}
	}
	return LoginRequest{}, errors.New("USER_NOT_FOUND")
}

func CreateInMemoryStorage() Storage {
	newStorage := InMemory{
		Members: []LoginRequest{
			LoginRequest{
				UserName:    "Hello123",
				Password:    "AXZ098",
				PhoneNumber: "6456546564",
			},
			LoginRequest{
				UserName:    "Bye456",
				Password:    "IOP678",
				PhoneNumber: "7987632156",
			},
		},
	}
	return newStorage
}

type Storage interface {
	Get(string) (LoginRequest, error)
}

func InitStorage() (Storage, error) {
	var env Environement
	if env.IsTest() {
		return CreateInMemoryStorage(), nil
	} else {
		return InMemory{}, errors.New("unsupported env")
	}
}
