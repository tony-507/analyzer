package main

import (
	"fmt"
	"os"

	"github.com/tony-507/analyzers/test"
)

func main() {
	fmt.Println("==========          Unit Test For analyzers          ==========")
	t := test.ConstructTestFlow()
	res := t.RunTests()
	if !res {
		fmt.Println("ERROR: test(s) failed. See logs above for more detail")
		os.Exit(1)
	}
}
