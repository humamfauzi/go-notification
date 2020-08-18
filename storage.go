package main

import "errors"

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

type Storage interface {
	Get(string) (UserLogin, error)
}

func InitStorage() (Storage, error) {
	var env Environement
	if env.IsTest() {
		return CreateInMemoryStorage(), nil
	} else {
		return InMemory{}, errors.New("unsupported env")
	}
}
