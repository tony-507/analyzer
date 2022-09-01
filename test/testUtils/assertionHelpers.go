package testUtils

import (
	"errors"
	"fmt"
	"reflect"
)

// This file stores methods for asserting test results

// Basic assert. Return bool and error

// Return true iff data type equal and value equal
func Assert_equal(a interface{}, b interface{}) (bool, error) {
	typeA := reflect.TypeOf(a)
	typeB := reflect.TypeOf(b)
	if typeA != typeB {
		errMsg := fmt.Sprintf("type not equal: expected %v, but got %v", typeB, typeA)
		err := errors.New(errMsg)
		return false, err
	}
	if a == b {
		return true, nil
	} else {
		va := reflect.ValueOf(a)
		vb := reflect.ValueOf(b)
		errMsg := fmt.Sprintf("value not equal: expected %v, but got %v", vb, va)
		err := errors.New(errMsg)
		return false, err
	}
}

// Return true if typeName of a equals b
func Assert_type_equal(a interface{}, b string) (bool, error) {
	typeName := reflect.TypeOf(a).String()
	if typeName == b {
		return true, nil
	} else {
		errMsg := fmt.Sprintf("value not equal: expected %s, but got %s", b, typeName)
		err := errors.New(errMsg)
		return false, err
	}
}
