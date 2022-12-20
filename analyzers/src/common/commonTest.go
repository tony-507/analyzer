package common

import (
	"errors"

	"github.com/tony-507/analyzers/src/testUtils"
)

func simpleBufTest() testUtils.Testcase {
	tc := testUtils.GetTestCase("simpleBufTest", 0)

	tc.Describe("Set data to simpleBuf", func(input interface{}) (interface{}, error) {
		buf := []byte{1, 2, 3}
		simpleBuf := MakeSimpleBuf(buf)
		simpleBuf.SetField("dummy", 100, false)
		return simpleBuf, nil
	})

	tc.Describe("Get data from simpleBuf", func(input interface{}) (interface{}, error) {
		simpleBuf, isBuf := input.(SimpleBuf)
		if !isBuf {
			panic("Input is not simpleBuf")
		}
		if field, hasField := simpleBuf.GetField("dummy"); hasField {
			if v, isInt := field.(int); isInt {
				testUtils.Assert_equal(v, 100)
			} else {
				panic("Data not int")
			}
		} else {
			panic("No data found")
		}
		return simpleBuf, nil
	})

	tc.Describe("Print simpleBuf data", func(input interface{}) (interface{}, error) {
		simpleBuf, isBuf := input.(SimpleBuf)
		if !isBuf {
			panic("Input is not simpleBuf")
		}
		testUtils.Assert_equal(simpleBuf.GetFieldAsString(), "dummy\n")
		testUtils.Assert_equal(simpleBuf.ToString(), "100\n")
		return nil, nil
	})

	return tc
}

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
	tmg.AddTest(simpleBufTest, []string{"common"})
	tmg.AddTest(readPcrTest, []string{"common"})
	tmg.AddTest(bsWriterTest, []string{"common"})

	t.AddSuite("unitTest", tmg)
}
