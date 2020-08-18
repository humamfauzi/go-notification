package main

import "errors"

type InMemory struct {
	Members []UserProfile
}

func (im InMemory) Get(userName string) (UserProfile, error) {
	for _, user := range im.Members {
		if user.UserDetail.Username == userName {
			return user, nil
		}
	}
	return UserProfile{}, errors.New("USER_NOT_FOUND")
}

func CreateInMemoryStorage() Storage {
	newStorage := InMemory{
		Members: []UserProfile{
			UserProfile{
				UserDetail{
					UserId:      "1",
					Username:    "Hello123",
					PhoneNumber: "6456546564",
				},
				UserCredential{
					UserId:   "1",
					Password: "AXZ098",
				},
			},
			UserProfile{
				UserDetail{
					UserId:      "2",
					Username:    "Hello456",
					PhoneNumber: "85665456464",
				},
				UserCredential{
					UserId:   "2",
					Password: "QWE1123",
				},
			},
		},
	}
	return newStorage
}

type Storage interface {
	Get(string) (UserProfile, error)
}

func InitStorage() (Storage, error) {
	var env Environement
	if env.IsTest() {
		return CreateInMemoryStorage(), nil
	} else {
		return InMemory{}, errors.New("unsupported env")
	}
}
