package ioUtils

import (
	"errors"

	"github.com/tony-507/analyzers/test/testUtils"
)

func initFileReaderTest() testUtils.Testcase {
	tc := testUtils.GetTestCase("initFileReaderTest", 0)

	tc.Describe("Set all parameters to be in use", func (input interface{}) (interface{}, error) {
		fr := GetReader("dummy")
		m_parameter := IOReaderParam{Fname: "dummy.ts", SkipCnt: 10, MaxInCnt: 10}

		fr.SetParameter(m_parameter)

		err := testUtils.Assert_equal(fr.ext, INPUT_TS)
		return nil, err
	})

	tc.Describe("Invalid input file format", func (input interface{}) (interface{}, error) {
		var err error
		defer func () {
			if _, ok := recover().(error); !ok {
				err = errors.New("FileReader.SetParameter does not panic on invalid input")
			}
		} ()
		fr := GetReader("dummy")
		m_parameter := IOReaderParam{Fname: "hello.abc"}

		fr.SetParameter(m_parameter)

		return nil, err
	})

	tc.Describe("Input file name with dot", func (input interface{}) (interface{}, error) {
		fr := GetReader("dummy")
		m_parameter := IOReaderParam{Fname: "hello.abc.ts"}

		fr.SetParameter(m_parameter)

		err := testUtils.Assert_equal(fr.ext, INPUT_TS)
		return nil, err
	})

	tc.Describe("Uninitialized maxInCnt", func (input interface{}) (interface{}, error) {
		fr := GetReader("dummy")
		m_parameter := IOReaderParam{Fname: "dummy.ts"}
		fr.SetParameter(m_parameter)

		err := testUtils.Assert_equal(fr.maxInCnt, -1)
		return nil, err
	})

	return tc
}

func AddIoUtilsTestSuite(t *testUtils.Tester) {
	tmg := testUtils.GetTestCaseMgr()

	tmg.AddTest(initFileReaderTest, []string{"ioUtils"})

	t.AddSuite("unitTest", tmg)
}
