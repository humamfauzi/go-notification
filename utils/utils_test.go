package utils

import (
	"testing"
)

func TestRandomStringId(t* testing.T) {
	preferedLength := 19
	result := RandomStringId("asd", preferedLength)
	if len(result) != 14 {
		t.Logf("want %v, have %v", len(result), preferedLength + 4)
	}
	t.Logf("%v", result)
	return
}