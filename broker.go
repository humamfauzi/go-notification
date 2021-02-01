package main

import (
	"fmt"
)

type Input struct {
	Topic string
}

func PrintInput (input Input) bool {
	fmt.Println(input)
	return true
}