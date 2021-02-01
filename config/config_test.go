package config 

import (
	"testing"
	"reflect"
)
func compare(t *testing.T, a, b interface{}) {
	comparison := `want %v; get %v`
	if a != b {
		t.Fatalf(comparison, a, b)
	}
}

func TestGetConfig(t *testing.T) {
	var config Config
	config.GetConfig("./test.config.json")
	comparison := `want %v; get %v`
	checkString := "astring"
	resultString := config["string"]
	compare(t, checkString, resultString)

	checkNumber := 1000
	resultNumber := config["number"]
	if checkString != resultString {
		t.Fatalf(comparison, checkNumber, resultNumber)
	}

	checkFloat := 123.321
	resultFloat := config["float"]
	if checkFloat != resultFloat {
		t.Fatalf(comparison, checkFloat, resultFloat)
	}

	checkBool := false
	resultBool := config["boolean"]
	if checkBool != resultBool {
		t.Fatalf(comparison, checkBool, resultBool)
	}

	checkArray := []interface{}{"asdf", 123, false}
	resultArray := config["array"]
	if reflect.DeepEqual(checkArray, resultArray) {
		t.Fatalf(comparison, checkArray, resultArray)
	}
	
	// checkNested := make(map[string]interface{})
	checkNested := make(map[string]interface{})
	checkNested["string"] = "astring"
	resultNested := config["nested"].(map[string]interface{})
	if reflect.DeepEqual(checkArray, resultNested) {
		t.Fatalf(comparison, checkNested, resultNested)
	}
}