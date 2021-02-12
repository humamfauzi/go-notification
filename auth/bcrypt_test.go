package auth

import (
	"testing"
)

func TestBcryptFunction(t *testing.T) {
	usernameEmail := "asd@asd.asd"
	passwordEmail := "secreat"
	result, err := BcryptConvertTo(usernameEmail, passwordEmail)
	if err != nil {
		t.Fatalf("Cannot convert to bcrypt hash %v", err)
	}
	combination := ComposeBcryptPassword(usernameEmail, passwordEmail)
	if ok := BcryptCheck([]byte(result), combination); !ok {
		t.Fatalf("Wrong password")
	}
	return
}