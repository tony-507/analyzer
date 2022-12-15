package ioUtils

import (
	"errors"

	"github.com/tony-507/analyzers/src/common"
	"github.com/tony-507/analyzers/test/testUtils"
)

// Helper
var TEST_OUT_DIR = testUtils.GetOutputDir() + "/test_output/"

func initFileReaderTest() testUtils.Testcase {
	tc := testUtils.GetTestCase("initFileReaderTest", 0)

	tc.Describe("Create file reader", func(input interface{}) (interface{}, error) {
		fr := GetReader("dummy")
		m_parameter := IOReaderParam{Source: SOURCE_FILE, FileInput: FileInputParam{Fname: "dummy.ts"}}

		fr.SetParameter(m_parameter)

		impl, isFileReader := fr.impl.(*fileReader)
		if !isFileReader {
			err := errors.New("File reader not created")
			return nil, err
		}

		err := testUtils.Assert_equal(impl.ext, INPUT_TS)
		return nil, err
	})

	tc.Describe("Invalid input file format", func(input interface{}) (interface{}, error) {
		var err error
		defer func() {
			if _, ok := recover().(error); !ok {
				err = errors.New("FileReader.SetParameter does not panic on invalid input")
			}
		}()
		fr := GetReader("dummy")
		m_parameter := IOReaderParam{Source: SOURCE_FILE, FileInput: FileInputParam{Fname: "hello.abc"}}

		fr.SetParameter(m_parameter)

		return nil, err
	})

	tc.Describe("Input file name with dot", func(input interface{}) (interface{}, error) {
		fr := GetReader("dummy")
		m_parameter := IOReaderParam{Source: SOURCE_FILE, FileInput: FileInputParam{Fname: "hello.abc.ts"}}

		fr.SetParameter(m_parameter)

		impl, isFileReader := fr.impl.(*fileReader)
		if !isFileReader {
			err := errors.New("Receiver is not created correctly")
			return nil, err
		}

		err := testUtils.Assert_equal(impl.ext, INPUT_TS)
		return nil, err
	})

	tc.Describe("Uninitialized maxInCnt", func(input interface{}) (interface{}, error) {
		fr := GetReader("dummy")
		m_parameter := IOReaderParam{Source: SOURCE_FILE, FileInput: FileInputParam{Fname: "dummy.ts"}}
		fr.SetParameter(m_parameter)

		err := testUtils.Assert_equal(fr.maxInCnt, -1)
		return nil, err
	})

	return tc
}

func readerDeliverUnitTest() testUtils.Testcase {
	tc := testUtils.GetTestCase("readerDeliverUnitTest", 0)

	tc.Describe("Deliver without additional settings", func(input interface{}) (interface{}, error) {
		ir := GetReader("dummy")
		m_parameter := IOReaderParam{Source: SOURCE_DUMMY}
		ir.SetParameter(m_parameter)

		ir.SetCallback(func(s string, unit common.CmUnit) {
			expected := common.MakeReqUnit(ir.name, common.FETCH_REQUEST)
			err := testUtils.Assert_obj_equal(expected, unit)
			if err != nil {
				panic(err)
			}
		})

		for i := 0; i < 5; i++ {
			ir.start()
		}

		ir.SetCallback(func(s string, unit common.CmUnit) {
			expected := common.MakeReqUnit(ir.name, common.EOS_REQUEST)
			err := testUtils.Assert_obj_equal(expected, unit)
			if err != nil {
				panic(err)
			}
		})
		ir.start()
		return nil, nil
	})

	tc.Describe("Deliver with skipping does not change behaviour", func(input interface{}) (interface{}, error) {
		ir := GetReader("dummy")
		m_parameter := IOReaderParam{Source: SOURCE_DUMMY, SkipCnt: 2}
		ir.SetParameter(m_parameter)

		ir.SetCallback(func(s string, unit common.CmUnit) {
			expected := common.MakeReqUnit(ir.name, common.FETCH_REQUEST)
			err := testUtils.Assert_obj_equal(expected, unit)
			if err != nil {
				panic(err)
			}
		})

		for i := 0; i < 5; i++ {
			ir.start()
		}

		ir.SetCallback(func(s string, unit common.CmUnit) {
			expected := common.MakeReqUnit(ir.name, common.EOS_REQUEST)
			err := testUtils.Assert_obj_equal(expected, unit)
			if err != nil {
				panic(err)
			}
		})

		ir.start()
		return nil, nil
	})

	tc.Describe("Deliver with max input count", func(input interface{}) (interface{}, error) {
		ir := GetReader("dummy")
		m_parameter := IOReaderParam{Source: SOURCE_DUMMY, MaxInCnt: 2}
		ir.SetParameter(m_parameter)

		ir.SetCallback(func(s string, unit common.CmUnit) {
			expected := common.MakeReqUnit(ir.name, common.FETCH_REQUEST)
			err := testUtils.Assert_obj_equal(expected, unit)
			if err != nil {
				panic(err)
			}
		})

		for i := 0; i < 2; i++ {
			ir.start()
		}

		ir.SetCallback(func(s string, unit common.CmUnit) {
			expected := common.MakeReqUnit(ir.name, common.EOS_REQUEST)
			err := testUtils.Assert_obj_equal(expected, unit)
			if err != nil {
				panic(err)
			}
		})

		ir.start()
		return nil, nil
	})

	return tc
}

func writerDeliverUnitTest() testUtils.Testcase {
	tc := testUtils.GetTestCase("writerDeliverUnitTest", 0)

	tc.Describe("Deliver without additional setting", func(input interface{}) (interface{}, error) {
		ow := GetOutputWriter("dummy")
		x := 0
		m_parameter := IOWriterParam{OutputType: OUTPUT_DUMMY, dummyOut: &x}
		ow.SetParameter(m_parameter)

		// Deliver some dummy units
		for i := 1; i < 5; i++ {
			unit := common.IOUnit{Buf: i}
			ow.DeliverUnit(unit)
		}

		err := testUtils.Assert_equal(x, 1234)
		return nil, err
	})

	return tc
}

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

func AddIoUtilsTestSuite(t *testUtils.Tester) {
	tmg := testUtils.GetTestCaseMgr()

	tmg.AddTest(initFileReaderTest, []string{"ioUtils"})
	tmg.AddTest(readerDeliverUnitTest, []string{"ioUtils"})
	tmg.AddTest(writerDeliverUnitTest, []string{"ioUtils"})
	tmg.AddTest(writerMultiThreadTest, []string{"ioUtils"})

	t.AddSuite("unitTest", tmg)
}
