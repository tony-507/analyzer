package test

import (
	"github.com/tony-507/analyzers/test/common"
	"github.com/tony-507/analyzers/test/testUtils"
)

func ConstructTestFlow() testUtils.Tester {
	// We may accept some filters here
	t := testUtils.GetTester()
	common.AddCommonSuite(&t)
	return t
}
