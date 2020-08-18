package main

import "os"

type Environement string

func (env Environement) getCurrentEnv() string {
	return os.Getenv("GO_ENV")
}
func (env Environement) IsTest() bool {
	return env.getCurrentEnv() == "test"
}
