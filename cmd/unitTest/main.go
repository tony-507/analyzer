package main

import (
	"fmt"

	"github.com/tony-507/analyzers/src/test"
)

func main() {
	fmt.Println("==========          Unit Test For analyzers          ==========")
	t := test.GetTester()
	test.AddCommonSuite(&t)
	t.RunTests()
}
