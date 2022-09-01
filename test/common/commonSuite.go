package common

import "github.com/tony-507/analyzers/test/testUtils"

func AddCommonSuite(t *testUtils.Tester) {
	tests := make([]testUtils.Testcase, 0)

	// We may add custom test filter here later
	tests = append(tests, readPcrTest())

	t.AddSuite("common", tests)
}
