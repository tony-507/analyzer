package test

import "fmt"

// This file declares several utils for testing

type INPUT_ENUM int

const (
	INPUT_RAW  INPUT_ENUM = 1 // Raw byte input
	INPUT_FILE INPUT_ENUM = 2 // File input, not supported yet
)

// Function prototype for a step. This follows nodeJs architecture
type runStep func(interface{}) (interface{}, error)

// A step for the test
// Basically we need an initialization step first, then perform what we want
// A clean up step may also be needed if necessary
type testStep struct {
	stepName string
	stepFunc runStep
}

// A test suite contains multiple test cases
type testSuite struct {
	tests     []testcase
	suiteName string
}

// A test case contains a test that follows the steps
type testcase struct {
	testSteps []testStep
	testName  string
}

// Helper struct for managing test cases
type testCaseMgr struct {
	testcases []testcase
	tags      [][]string // Carry tags for tests. First index is the tag, second is case index
}

// The struct that actually runs tests
type Tester struct {
	suites []testSuite
}

// Add a test step
func (tc *testcase) describe(stepName string, stepFunc runStep) {
	step := testStep{stepName: stepName, stepFunc: stepFunc}
	tc.testSteps = append(tc.testSteps, step)
}

// Add a test case
func (tmg *testCaseMgr) addTest(flow testcase, tags []string) {
	tmg.testcases = append(tmg.testcases, flow)
}

func (t *Tester) addSuite(suiteName string, tests []testcase) {
	ts := testSuite{suiteName: suiteName, tests: tests}
	t.suites = append(t.suites, ts)
}

func GetTester() Tester {
	t := Tester{suites: make([]testSuite, 0)}
	return t
}

// Function that runs all selected tests
func (t *Tester) RunTests() {
	for _, suite := range t.suites {
		fmt.Println("\nSuite:", suite.suiteName)
		for _, test := range suite.tests {
			res := true
			outMsg := ""
			errMsg := "" // Prevent segfault
			curStep := ""
			// Dummy input for initialization at first step
			rv, err := test.testSteps[0].stepFunc(nil)
			if err != nil {
				errMsg = err.Error()
				res = false
			} else {
				// Continue only if first step is OK
				for idx := 1; idx < len(test.testSteps); idx++ {
					curStep = test.testSteps[idx].stepName
					rv, err = test.testSteps[idx].stepFunc(rv)
					if err != nil {
						errMsg = err.Error()
						res = false
						break
					}
				}
			}

			if !res {
				outMsg = fmt.Sprintf("%s: %t, failure reason: %s at step \"%s\"", test.testName, res, errMsg, curStep)
			} else {
				outMsg = fmt.Sprintf("%s: %t", test.testName, res)
			}
			fmt.Println(outMsg)
		}
	}
}
