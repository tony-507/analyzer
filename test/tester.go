package test

import (
	"github.com/tony-507/analyzers/src/common"
	"github.com/tony-507/analyzers/src/worker"
	"github.com/tony-507/analyzers/test/testUtils"
)

func ConstructTestFlow() testUtils.Tester {
	// We may accept some filters here
	t := testUtils.GetTester()
	worker.AddUnitTestSuite(&t)
	common.AddUnitTestSuite(&t)
	return t
}
