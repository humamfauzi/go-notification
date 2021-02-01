package config

import (
	"io/ioutil"
	"encoding/json"
	"strings"
	"os"
)

type Config map[string]interface{}

type ConfigDatabase struct {
	Url string
	Port int
	Username string
	Password string
}

type ConfigAuth struct {
	Algorithm string
}

func (c *Config) GetConfig(directory string) (error) {
	configBody, err := ioutil.ReadFile(directory)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(configBody, c); err != nil {
		return err
	}
	return nil
}

func (c Config) Goto(path string, defaultValue interface{}) interface{} {
	pathArray := strings.Split(path, ".")
	lastPath := pathArray[len(pathArray)-1]
	pathArray = pathArray[:len(pathArray)-1]
	currentDepth := c
	for _, path := range pathArray {
		var ok bool
		currentDepth, ok = c[path].(map[string]interface{})
		if !ok {
			return defaultValue
		}
	}
	return currentDepth[lastPath]
}

func MapToStruct(jsonMap map[string]interface{}, bufferType interface{}) {
	jsonByte, _ := json.Marshal(jsonMap)
	json.Unmarshal(jsonByte, &bufferType)
}
