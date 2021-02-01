package auth

import (
	"testing"
	"time"
	"github.com/dgrijalva/jwt-go"
	"errors"
)

const (
	TEST_SECRET = "dsN(Dt2s,@K.Y#]&"
)

func keyFunc(token *jwt.Token) (interface{}, error) {
	if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
		return nil, errors.New("There was an error")
	}
	return []byte(TEST_SECRET), nil
}

func TestCreateToken(t *testing.T) {
	header := jwt.MapClaims{
		"header1": true,
		"header2": "someone",
		"exp": time.Now().Add(time.Minute * 15).Unix(),
	}
	_, err := CreateToken(header, []byte(TEST_SECRET))
	if err != nil {
		t.Fatalf("token fail")
	}
}

func TestVerifyToken(t *testing.T) {
	header := jwt.MapClaims{
		"test": true,
		"exp": time.Now().Add(time.Minute * 10).Unix(),
	}
	result, _ := CreateToken(header, []byte(TEST_SECRET))
	bearerToken := "Bearer " + result
	ok := VerifyToken(bearerToken, keyFunc)
	if !ok {
		t.Fatalf("Error")
	}
}

func TestCheckTokenExpiry(t *testing.T) {
	header := jwt.MapClaims{
		"header1": true,
		"header2": "someone",
		"exp": time.Now().Add(time.Minute * 15).Unix(),
	}
	if ok := CheckTokenExpiry(header["exp"]); !ok {
		t.Fatalf("Should pass")
	}

	header = jwt.MapClaims{
		"exp": time.Now().Add(time.Minute * -15).Unix(),
	}
	if ok := CheckTokenExpiry(header["exp"]); ok {
		t.Fatalf("Should fail")
	}
}