package testUtils

import (
	"errors"
	"fmt"
	"reflect"
)

// This file stores methods for assertions

func assertException(errMsg string) error {
	return errors.New(errMsg)
}

// Return true iff data type equal and value equal
func Assert_equal(a interface{}, b interface{}) error {
	typeA := reflect.TypeOf(a)
	typeB := reflect.TypeOf(b)
	if typeA != typeB {
		return assertException(fmt.Sprintf("type not equal: expected %v, but got %v", typeB, typeA))
	}
	if a == b {
		return nil
	} else {
		va := reflect.ValueOf(a)
		vb := reflect.ValueOf(b)
		return assertException(fmt.Sprintf("value not equal: expected %v, but got %v", vb, va))
	}
}

// Return true if typeName of a equals b
func Assert_type_equal(a interface{}, b string) error {
	typeName := reflect.TypeOf(a).String()
	if typeName == b {
		return nil
	} else {
		return assertException(fmt.Sprintf("value not equal: expected %s, but got %s", b, typeName))
	}
}

// Return true if every member of the struct is equal
func Assert_obj_equal(a interface{}, b interface{}) error {
	// Use reflect.DeepEqual
	// TODO For better performance, please implement equals function in struct-wise manner
	isEqual := reflect.DeepEqual(a, b)
	if !isEqual {
		return assertException(fmt.Sprintf("Objects not equal:\nExpecting %v\nGot %v\n", a, b))
	}
	return nil
}
