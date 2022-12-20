package main

import (
	"fmt"
	"os"

	"github.com/tony-507/analyzers/src/avContainer"
	"github.com/tony-507/analyzers/src/common"
	"github.com/tony-507/analyzers/src/ioUtils"
	"github.com/tony-507/analyzers/src/logs"
	"github.com/tony-507/analyzers/src/testUtils"
	"github.com/tony-507/analyzers/src/worker"
)

func main() {
	logs.SetProperty("level", "disabled")
	logs.SetProperty("format", "%t [%n] [%l] %s")

	t := testUtils.GetTester()

	// Unit tests
	common.AddUnitTestSuite(&t)
	ioUtils.AddIoUtilsTestSuite(&t)
	avContainer.AddUnitTestSuite(&t)
	worker.AddUnitTestSuite(&t)

	res := t.RunTests(os.Args)
	if !res {
		fmt.Println("ERROR: test(s) failed. See logs above for more detail")
		os.Exit(1)
	}
}
