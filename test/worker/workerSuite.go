package worker

import (
	"github.com/tony-507/analyzers/test/testUtils"
)

func AddWorkerSuite(t *testUtils.Tester) {
	tests := make([]testUtils.Testcase, 0)

	// We may add custom test filter here later
	tests = append(tests, pluginUnitTest())

	t.AddSuite("worker", tests)
}
