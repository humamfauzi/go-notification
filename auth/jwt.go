package auth

import (
	"github.com/dgrijalva/jwt-go"
	"strings"
	"errors"
	"time"
	"fmt"
)

const (
	/**
		Should outsorced to utils for security purpose but for now this would do
	*/
	AUTH_TOKEN_SECRET = "=kVsu2{G9'{'K<>d"
)

func GetAuthSecret() []byte {
	return []byte(AUTH_TOKEN_SECRET)
}

func CreateToken(mapClaims jwt.MapClaims, secret []byte) (string, error) {
	claims := jwt.NewWithClaims(jwt.SigningMethodHS256, mapClaims)
	token, err := claims.SignedString(secret)
	if err != nil {
		return "", err
	}
	return token, nil
}

func KeyFunction(token *jwt.Token) (interface{}, error) {
	if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
		return nil, errors.New("asdlkj")
	}
	if ok := CheckTokenExpiry(token.Header["exp"]); !ok {
		return nil, errors.New("Token Expired")
	}
	return GetAuthSecret(), nil
}

func CheckTokenExpiry(expiry interface{}) bool {
	var unixTime int64
	switch expiry.(type) {
	case float64:
		unixTime = int64(expiry.(float64))
	default:
		unixTime = expiry.(int64)
	}
	unixTime, ok := expiry.(int64)
	if !ok {
		unixTime = int64(expiry.(float64))
	}
	tokenTime := time.Unix(unixTime, 10)
	currentTime := time.Now()
	if currentTime.After(tokenTime) {
		return false
	}
	return true
}

func VerifyToken(authToken string, keyFunction func(token *jwt.Token) (interface{}, error)) bool {
	tokenBearer := ParseBearer(authToken)
	token, err := jwt.Parse(tokenBearer, keyFunction)	
	if err != nil {
		return false
	}
	claims, ok := token.Claims.(jwt.MapClaims);
	if !ok {
		return false
	}
	if ok = CheckTokenExpiry(claims["exp"]); !ok {
		fmt.Println(ok)
		return false
	}
	if token.Valid {
		return true
	} else {
		return false
	}
}

func BearerToken(jwtToken []byte) string {
	return "Bearer " + string(jwtToken)
}

func ParseBearer(receivedToken string) string {
	splitToken := strings.Split(receivedToken, " ")
	return splitToken[1]
}