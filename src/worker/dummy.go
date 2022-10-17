package worker

import (
	"errors"
	"time"

	"github.com/tony-507/analyzers/src/common"
	"github.com/tony-507/analyzers/src/logs"
	"github.com/tony-507/analyzers/src/resources"
)

// This file stores some dummy struct for testing
type DummyPlugin struct {
	logger   logs.Log
	inCnt    int
	fetchCnt int
	callback *Worker
	name     string
	role     int // 0 represents a root, 1 represents non-root
}

func (dp *DummyPlugin) SetParameter(m_parameter interface{}) {}

func (pg *DummyPlugin) SetResource(resourceLoader *resources.ResourceLoader) {
	// pg.impl.SetResource(resourceLoader)
}

func (dp *DummyPlugin) StartSequence() {
	if dp.role == 0 {
		unit := common.IOUnit{Buf: 20, IoType: 0, Id: 0}
		go dp.DeliverUnit(unit)
	}
}

func (dp *DummyPlugin) EndSequence() {
	eosUnit := common.MakeReqUnit(dp.name, common.EOS_REQUEST)
	dp.callback.PostRequest(dp.name, eosUnit)
}

func (dp *DummyPlugin) DeliverUnit(unit common.CmUnit) (bool, error) {
	// Ensure correct order of calling by suspending worker thread
	if dp.role == 0 {
		time.Sleep(time.Second)
	}

	buf, isInt := unit.GetBuf().(int)
	if !isInt {
		err := errors.New("buf is not int")
		return false, err
	}
	dp.inCnt += buf

	if buf > 10 {
		reqUnit := common.MakeReqUnit(dp.name, common.FETCH_REQUEST)
		dp.callback.PostRequest(dp.name, reqUnit)
	}

	if dp.role == 0 {
		eosUnit := common.MakeReqUnit(dp.name, common.EOS_REQUEST)
		dp.callback.PostRequest(dp.name, eosUnit)
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

func GetDummyPlugin(name string, isRoot int) DummyPlugin {
	return DummyPlugin{logger: logs.CreateLogger("Dummy"), inCnt: 0, fetchCnt: 0, name: name, role: isRoot}
}
