package test

import (
	"errors"
	"fmt"

	"github.com/tony-507/analyzers/src/common"
)

// This is a file containing tests for common module

func readPcrTest() testcase {
	tc := testcase{testName: "Read PCR test", testSteps: make([]testStep, 0)}

	tc.describe("Initialization", func(input interface{}) (interface{}, error) {
		buf := []byte{0x0e, 0x26, 0xe0, 0x33, 0x7e, 0x11}
		r := common.GetBufferReader(buf)
		return r, nil
	})

	tc.describe("Read 33 bit base", func(input interface{}) (interface{}, error) {
		r, isReader := input.(common.BsReader)
		if !isReader {
			err := errors.New("Reader not passed to next step")
			return nil, err
		}
		val := r.ReadBits(33)
		if !assert_equal(val, 474857574) {
			outMsg := fmt.Sprintf("Expected 474857574, but got %d", val)
			err := errors.New(outMsg)
			return nil, err
		}
		return r, nil
	})

	tc.describe("Read reserved bits", func(input interface{}) (interface{}, error) {
		r, isReader := input.(common.BsReader)
		if !isReader {
			err := errors.New("Reader not passed to next step")
			return nil, err
		}
		val := r.ReadBits(6)
		if !assert_equal(val, 63) {
			outMsg := fmt.Sprintf("Expected 63, but got %d", val)
			err := errors.New(outMsg)
			return nil, err
		}
		return r, nil
	})

	tc.describe("Read 9 bits extension", func(input interface{}) (interface{}, error) {
		r, isReader := input.(common.BsReader)
		if !isReader {
			err := errors.New("Reader not passed to next step")
			return nil, err
		}
		val := r.ReadBits(9)
		if !assert_equal(val, 17) {
			outMsg := fmt.Sprintf("Expected 17, but got %d", val)
			err := errors.New(outMsg)
			return nil, err
		}
		return r, nil
	})
	return tc
}

func AddCommonSuite(t *Tester) {
	tests := make([]testcase, 0)

	// We may add custom test filter here later
	tests = append(tests, readPcrTest())

	t.addSuite("common", tests)
}
