package auth

import (
	"golang.org/x/crypto/bcrypt"
)

const SALT = "s+rS?:wk&FeXN88EW"

func ComposeBcryptPassword(username, password string) []byte {
	combination := username + ":" + password + ":" + SALT
	return []byte(combination)
}

func BcryptConvertTo(username, password string) (string, error) {
	combination := ComposeBcryptPassword(username, password)
	hashedPassword, err := bcrypt.GenerateFromPassword(combination, bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

func BcryptCheck(hashedPassword, password []byte) bool {
	if err := bcrypt.CompareHashAndPassword(hashedPassword, password); err != nil {
		return false
	}
	return true
}