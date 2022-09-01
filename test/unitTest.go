package main

import "fmt"

func main() {
	fmt.Println("==========          Unit Test For analyzers          ==========")
	t := tester{suites: make([]testSuite, 0)}
	addCommonSuite(&t)
	t.runTests()
}
