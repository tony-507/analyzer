package worker

import (
	"errors"

	"github.com/tony-507/analyzers/src/worker"

	"github.com/tony-507/analyzers/src/common"
	"github.com/tony-507/analyzers/test/testUtils"
)

func pluginUnitTest() testUtils.Testcase {
	tc := testUtils.GetTestCase("Plugin unit test")

	tc.Describe("Initialization", func(input interface{}) (interface{}, error) {
		return "Dummy_1", nil
	})

	tc.Describe("Deduce plugin type", func(input interface{}) (interface{}, error) {
		cellName, isStr := input.(string)
		var err error
		if !isStr {
			err := errors.New("cellName is not a string")
			return nil, err
		}
		pg := worker.GetPluginByName(cellName)
		_, err = testUtils.Assert_type_equal(pg.Work, "worker.DummyPlugin")
		if err != nil {
			return nil, err
		}
		return pg, nil
	})

	tc.Describe("Run DeliverUnit interface", func(input interface{}) (interface{}, error) {
		pg, isPlugin := input.(worker.Plugin)
		if !isPlugin {
			err := errors.New("input is not a plugin")
			return nil, err
		}
		inUnit := common.IOUnit{Buf: 2, IoType: 0, Id: 0}
		_, err := pg.DeliverUnit(inUnit)
		if err != nil {
			return nil, err
		}
		return pg, nil
	})

	tc.Describe("Run FetchUnit interface", func(input interface{}) (interface{}, error) {
		pg, isPlugin := input.(worker.Plugin)
		if !isPlugin {
			err := errors.New("input is not a plugin")
			return nil, err
		}
		unit, err := pg.FetchUnit()
		if err != nil {
			return nil, err
		}

		rv, isInt := unit.GetBuf().(int)
		if !isInt {
			err := errors.New("return value is not an integer")
			return nil, err
		}
		if rv != 2 {
			err := errors.New("return value not equal to 2")
			return nil, err
		}
		return nil, nil
	})

	return tc
}
