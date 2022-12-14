package ioUtils

import (
	"github.com/tony-507/analyzers/src/common"
	"github.com/tony-507/analyzers/test/testUtils"
)

// Helper
var TEST_OUT_DIR = testUtils.GetOutputDir() + "/test_output/"

func writerMultiThreadTest() testUtils.Testcase {
	tc := testUtils.GetTestCase("writerMultiThreadTest", 0)

	tc.Describe("Setup file writer", func(input interface{}) (interface{}, error) {
		fw := GetFileWriter()
		param := IOWriterParam{FileOutput: FileOutputParam{OutFolder: TEST_OUT_DIR}}
		fw.setup(param)

		return fw, nil
	})

	tc.Describe("Start two raw data processors", func(input interface{}) (interface{}, error) {
		fw, _ := input.(*FileWriter)

		rawUnit := common.IOUnit{Buf: 1, IoType: 3, Id: 5}
		rawUnit2 := common.IOUnit{Buf: 1, IoType: 3, Id: 2}
		fw.processUnit(rawUnit)
		fw.processUnit(rawUnit2)

		return fw, nil
	})

	tc.Describe("Stop file writer", func(input interface{}) (interface{}, error) {
		fw, _ := input.(*FileWriter)
		fw.stop()
		return nil, nil
	})

	return tc
}

func AddIoUtilFileTestSuite(t *testUtils.Tester) {
	tmg := testUtils.GetTestCaseMgr()

	tmg.AddTest(writerMultiThreadTest, []string{"ioUtils"})

	t.AddSuite("unitTest", tmg)
}
