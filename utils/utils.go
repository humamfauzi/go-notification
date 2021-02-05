package utils

import (
	"math/rand"
	"strings"
	"time"
)

const (
	randomAlphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	randomAlphaNumeric = randomAlphabet + "1234567890"
	randomAlphaNumericSpecialChar = randomAlphaNumeric + "~`!@#$%^&*()_+-=<>?"
)

func RandomStringId(prefix string, length int) string {
	randomString := make([]string, length)
	for i:=0; i < length; i++ {
		randomString[i] = pickRandomCharacter("alpha")
	}
	
	return prefix +"/"+ strings.Join(randomString, "") 
}

func pickRandomCharacter(collection string) string {
	rand.Seed(time.Now().UnixNano())
	switch collection {
	case "alphanum":
		return string(randomAlphaNumeric[rand.Intn(len(randomAlphaNumeric))])
	case "alphanumspecial":
		return string(randomAlphaNumericSpecialChar[rand.Intn(len(randomAlphaNumericSpecialChar))])
	default:
		return string(randomAlphabet[rand.Intn(len(randomAlphabet))])
	}
}