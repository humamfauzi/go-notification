package main

import (
	"testing"
)

func TestPrintInput(t *testing.T) {
	name := Input{"Lala"}
	want := true
	result := PrintInput(name)
	if result != want {
		t.Fatalf(`Want %v get %v`, want, result)
	}
};