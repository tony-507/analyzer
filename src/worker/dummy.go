package worker

import (
	"errors"

	"github.com/tony-507/analyzers/src/common"
	"github.com/tony-507/analyzers/src/resources"
)

// This file stores some dummy struct for testing
type DummyPlugin struct {
	inCnt    int
	fetchCnt int
	callback *Worker
	name     string
}

func (dp *DummyPlugin) SetParameter(m_parameter interface{}) {}

func (pg *DummyPlugin) SetResource(resourceLoader *resources.ResourceLoader) {
	// pg.impl.SetResource(resourceLoader)
}

func (dp *DummyPlugin) StartSequence() {}

func (dp *DummyPlugin) EndSequence() {}

func (dp *DummyPlugin) DeliverUnit(unit common.CmUnit) (bool, error) {
	buf, isInt := unit.GetBuf().(int)
	if !isInt {
		err := errors.New("buf is not int")
		return false, err
	}
	dp.inCnt += buf

	if buf > 10 {
		reqUnit := common.MakeReqUnit(nil, common.FETCH_REQUEST)
		dp.callback.PostRequest(dp.name, reqUnit)
	}

	return true, nil
}

func (dp *DummyPlugin) FetchUnit() common.CmUnit {
	rv := common.IOUnit{Buf: dp.inCnt*10 + dp.fetchCnt, IoType: 0, Id: -1}
	dp.fetchCnt += 1
	return rv
}

func (dp *DummyPlugin) SetCallback(callback *Worker) {
	dp.callback = callback
}

func GetDummyPlugin(name string) DummyPlugin {
	return DummyPlugin{inCnt: 0, fetchCnt: 0, name: name}
}
