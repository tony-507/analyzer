package worker

import (
	"errors"

	"github.com/tony-507/analyzers/src/common"
)

// This file stores some dummy struct for testing
type DummyPlugin struct {
	cnt int
}

func (dp *DummyPlugin) DeliverUnit(unit common.CmUnit) (bool, error) {
	buf, isInt := unit.GetBuf().(int)
	if !isInt {
		err := errors.New("buf is not int")
		return false, err
	}
	dp.cnt += buf
	return true, nil
}

func (dp *DummyPlugin) FetchUnit() (common.CmUnit, error) {
	rv := common.IOUnit{Buf: dp.cnt, IoType: 0, Id: -1}
	return rv, nil
}

func GetDummyPlugin() DummyPlugin {
	return DummyPlugin{cnt: 0}
}
