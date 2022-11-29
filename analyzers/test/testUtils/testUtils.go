package testUtils

import (
	"fmt"
	"time"
)

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
	timeout   time.Duration
}

// Helper struct for managing test cases
type TestCaseMgr struct {
	setups []testSetup
	tags   [][]string // Carry tags for tests. First index is the tag, second is test index
}

// The struct that actually runs tests
type Tester struct {
	suites []TestSuite
	// Used for debugging on test failure
	curStep string
	errMsg  string
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

func GetTestCase(name string, timeout int) Testcase {
	tc := Testcase{testName: name, timeout: time.Duration(timeout) * time.Second, TestSteps: make([]TestStep, 0)}
	return tc
}

func GetTestCaseMgr() TestCaseMgr {
	return TestCaseMgr{setups: make([]testSetup, 0), tags: make([][]string, 0)}
}

func GetTester() Tester {
	t := Tester{suites: make([]TestSuite, 0)}
	return t
}

// If a test panics, recover and go to the next one
func (t *Tester) runNextIfPanic() {
	if err, ok := recover().(error); ok {
		t.errMsg = err.Error()
	}
}

// Function that runs a test setup
// It is implemented separately to allow recovery of testing after a test fails
func (t *Tester) _runTestSetup(test Testcase, pair string) bool {
	// defer t.runNextIfPanic()
	res := true

	fmt.Printf("\tRunning %s\n", pair)

	// Dummy input for initialization at first step
	t.curStep = test.TestSteps[0].stepName
	rv, err := test.TestSteps[0].stepFunc(nil)
	if err != nil {
		t.errMsg = err.Error()
		res = false
	} else {
		// Continue only if first step is OK
		for idx := 1; idx < len(test.TestSteps); idx++ {
			t.curStep = test.TestSteps[idx].stepName
			rv, err = test.TestSteps[idx].stepFunc(rv)
			if err != nil {
				t.errMsg = err.Error()
				res = false
				break
			}
		}
	}
	return res
}

// Function that runs all selected tests
func (t *Tester) RunTests() bool {
	testCh := make(chan bool, 1)
	isPass := true

	for _, suite := range t.suites {
		// Test statistics
		runTotal := 0
		passTotal := 0
		for _, setup := range suite.tests {
			outMsg := ""
			test := setup()
			pair := fmt.Sprintf("%s.%s", suite.suiteName, test.testName)
			// Set default timeout
			if test.timeout == 0 {
				test.timeout = 2 * time.Second
			}
			go func() {
				res := t._runTestSetup(test, pair)
				testCh <- res
			}()

			res := false
			select {
			case rv := <-testCh:
				res = rv
			case <-time.After(test.timeout):
				t.errMsg = "Test timeout"
			}

			if !res {
				outMsg = fmt.Sprintf("%s\n\t[FAILED] %s at step \"%s\"\n", t.errMsg, pair, t.curStep)
				isPass = false
			} else {
				outMsg = fmt.Sprintf("[PASS] %s\n", pair)
				passTotal += 1
			}

			runTotal += 1
			fmt.Printf("\t%s\n", outMsg)

			t.errMsg = ""
			t.curStep = ""
		}
		statMsg := fmt.Sprintf("[Suite %s] Executed: %d, passed: %d\n", suite.suiteName, runTotal, passTotal)
		fmt.Printf("\t%s\n", statMsg)
	}

	return isPass
}
