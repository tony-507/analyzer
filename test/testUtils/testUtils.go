package testUtils

import "fmt"

// This file declares several utils for testing

type INPUT_ENUM int

const (
	INPUT_RAW  INPUT_ENUM = 1 // Raw byte input
	INPUT_FILE INPUT_ENUM = 2 // File input, not supported yet
)

// Function prototype for a step. This follows nodeJs architecture
type runStep func(interface{}) (interface{}, error)

// Function prototype for a testcase. We store a function so that resources are consumed when we need to run the test
type testSetup func() Testcase

// A step for the test
// Basically we need an initialization step first, then perform what we want
// A clean up step may also be needed if necessary
type TestStep struct {
	stepName string
	stepFunc runStep
}

// A test suite contains multiple test cases
type TestSuite struct {
	tests     []testSetup
	suiteName string
}

// A test case contains a test that follows the steps
type Testcase struct {
	TestSteps []TestStep
	testName  string
}

// Helper struct for managing test cases
type TestCaseMgr struct {
	setups []testSetup
	tags   [][]string // Carry tags for tests. First index is the tag, second is test index
}

// The struct that actually runs tests
type Tester struct {
	suites []TestSuite
}

// Add a test step
func (tc *Testcase) Describe(stepName string, stepFunc runStep) {
	step := TestStep{stepName: stepName, stepFunc: stepFunc}
	tc.TestSteps = append(tc.TestSteps, step)
}

// Add a test case
func (tmg *TestCaseMgr) AddTest(flow testSetup, tags []string) {
	tmg.setups = append(tmg.setups, flow)
}

func (t *Tester) AddSuite(suiteName string, tmg TestCaseMgr) {
	// If suite already exists, append tests to the suite
	for idx, suite := range t.suites {
		if suite.suiteName == suiteName {
			t.suites[idx].tests = append(t.suites[idx].tests, tmg.setups...)
			return
		}
	}
	// If new suite, create it
	ts := TestSuite{suiteName: suiteName, tests: tmg.setups}
	t.suites = append(t.suites, ts)
}

func GetTestCase(name string) Testcase {
	tc := Testcase{testName: name, TestSteps: make([]TestStep, 0)}
	return tc
}

func GetTestCaseMgr() TestCaseMgr {
	return TestCaseMgr{setups: make([]testSetup, 0), tags: make([][]string, 0)}
}

func GetTester() Tester {
	t := Tester{suites: make([]TestSuite, 0)}
	return t
}

// Function that runs all selected tests
func (t *Tester) RunTests() {
	for _, suite := range t.suites {
		for _, setup := range suite.tests {
			res := true
			outMsg := ""
			errMsg := "" // Prevent segfault
			// Set up the test
			test := setup()
			pair := fmt.Sprintf("%s.%s", suite.suiteName, test.testName)
			fmt.Printf("Running %s\n", pair)

			curStep := ""
			// Dummy input for initialization at first step
			rv, err := test.TestSteps[0].stepFunc(nil)
			if err != nil {
				errMsg = err.Error()
				res = false
			} else {
				// Continue only if first step is OK
				for idx := 1; idx < len(test.TestSteps); idx++ {
					curStep = test.TestSteps[idx].stepName
					rv, err = test.TestSteps[idx].stepFunc(rv)
					if err != nil {
						errMsg = err.Error()
						res = false
						break
					}
				}
			}

			if !res {
				outMsg = fmt.Sprintf("[FAILED] %s: %s at step \"%s\"\n", pair, errMsg, curStep)
			} else {
				outMsg = fmt.Sprintf("[PASS] %s\n", pair)
			}
			fmt.Println(outMsg)
		}
	}
}
