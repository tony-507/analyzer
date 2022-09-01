package main

import (
	"reflect"
)

// This file stores methods for asserting test results

// Basic assert
// Return true iff data type equal and value equal
func assert_equal(a interface{}, b interface{}) bool {
	if reflect.TypeOf(a) != reflect.TypeOf(b) {
		return false
	}
	return a == b
}
