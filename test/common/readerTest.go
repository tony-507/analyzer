package common

import (
	"errors"

	"github.com/tony-507/analyzers/src/common"
	"github.com/tony-507/analyzers/test/testUtils"
)

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
		res, err := testUtils.Assert_equal(val, 474857574)
		if !res {
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
		res, err := testUtils.Assert_equal(val, 63)
		if !res {
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
		res, err := testUtils.Assert_equal(val, 17)
		if !res {
			return nil, err
		}
		return r, nil
	})
	return tc
}
