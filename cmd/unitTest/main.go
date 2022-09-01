package main

import (
	"fmt"

	"github.com/tony-507/analyzers/test"
)

func main() {
	fmt.Println("==========          Unit Test For analyzers          ==========")
	t := test.ConstructTestFlow()
	t.RunTests()
}
