package common

import (
	"errors"
	"fmt"

	"github.com/tony-507/analyzers/src/common"
	"github.com/tony-507/analyzers/test/testUtils"
)

// This is a file containing tests for common module

func readPcrTest() testUtils.Testcase {
	tc := testUtils.GetTestCase("Reader test")

	tc.Describe("Initialization", func(input interface{}) (interface{}, error) {
		buf := []byte{0x0e, 0x26, 0xe0, 0x33, 0x7e, 0x11}
		r := common.GetBufferReader(buf)
		return r, nil
	})

	tc.Describe("Read 33 bits", func(input interface{}) (interface{}, error) {
		r, isReader := input.(common.BsReader)
		if !isReader {
			err := errors.New("Reader not passed to next step")
			return nil, err
		}
		val := r.ReadBits(33)
		if !testUtils.Assert_equal(val, 474857574) {
			outMsg := fmt.Sprintf("Expected 474857574, but got %d", val)
			err := errors.New(outMsg)
			return nil, err
		}
		return r, nil
	})

	tc.Describe("Read 6 bits", func(input interface{}) (interface{}, error) {
		r, isReader := input.(common.BsReader)
		if !isReader {
			err := errors.New("Reader not passed to next step")
			return nil, err
		}
		val := r.ReadBits(6)
		if !testUtils.Assert_equal(val, 63) {
			outMsg := fmt.Sprintf("Expected 63, but got %d", val)
			err := errors.New(outMsg)
			return nil, err
		}
		return r, nil
	})

	tc.Describe("Read 9 bits", func(input interface{}) (interface{}, error) {
		r, isReader := input.(common.BsReader)
		if !isReader {
			err := errors.New("Reader not passed to next step")
			return nil, err
		}
		val := r.ReadBits(9)
		if !testUtils.Assert_equal(val, 17) {
			outMsg := fmt.Sprintf("Expected 17, but got %d", val)
			err := errors.New(outMsg)
			return nil, err
		}
		return r, nil
	})
	return tc
}
