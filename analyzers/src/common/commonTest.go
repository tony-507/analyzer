package common

import (
	"errors"

	"github.com/tony-507/analyzers/test/testUtils"
)

func readPcrTest() testUtils.Testcase {
	tc := testUtils.GetTestCase("readerTest", 0)

	tc.Describe("Initialization", func(input interface{}) (interface{}, error) {
		buf := []byte{0x0e, 0x26, 0xe0, 0x33, 0x7e, 0x11}
		r := GetBufferReader(buf)
		return r, nil
	})

	tc.Describe("Read 33 bits", func(input interface{}) (interface{}, error) {
		r, isReader := input.(BsReader)
		if !isReader {
			err := errors.New("Reader not passed to next step")
			return nil, err
		}
		val := r.ReadBits(33)
		err := testUtils.Assert_equal(val, 474857574)
		if err != nil {
			return nil, err
		}
		return r, nil
	})

	tc.Describe("Read 6 bits", func(input interface{}) (interface{}, error) {
		r, isReader := input.(BsReader)
		if !isReader {
			err := errors.New("Reader not passed to next step")
			return nil, err
		}
		val := r.ReadBits(6)
		err := testUtils.Assert_equal(val, 63)
		if err != nil {
			return nil, err
		}
		return r, nil
	})

	tc.Describe("Read 9 bits", func(input interface{}) (interface{}, error) {
		r, isReader := input.(BsReader)
		if !isReader {
			err := errors.New("Reader not passed to next step")
			return nil, err
		}
		val := r.ReadBits(9)
		err := testUtils.Assert_equal(val, 17)
		if err != nil {
			return nil, err
		}
		return r, nil
	})
	return tc
}

func bsWriterTest() testUtils.Testcase {
	tc := testUtils.GetTestCase("bsWriterTest", 0)

	tc.Describe("Basic bit writing", func(input interface{}) (interface{}, error) {
		writer := GetBufferWriter(3)

		writer.writeBits(0x47, 8)
		writer.writeBits(0, 1)
		writer.writeBits(0, 1)
		writer.writeBits(0, 1)
		writer.writeBits(33, 13)

		expected := []byte{0x47, 0x00, 0x21}
		err := testUtils.Assert_obj_equal(expected, writer.GetBuf())

		return nil, err
	})

	return tc
}

func AddUnitTestSuite(t *testUtils.Tester) {
	tmg := testUtils.GetTestCaseMgr()

	// We may add custom test filter here later

	// common
	tmg.AddTest(readPcrTest, []string{"common"})
	tmg.AddTest(bsWriterTest, []string{"common"})

	t.AddSuite("unitTest", tmg)
}
