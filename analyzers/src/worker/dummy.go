package worker

import (
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
	callback common.RequestHandler
	name     string
	role     int // 0 represents a root, 1 represents non-root
}

func (dp *DummyPlugin) SetParameter(m_parameter string) {}

func (pg *DummyPlugin) SetResource(resourceLoader *resources.ResourceLoader) {
	// pg.impl.SetResource(resourceLoader)
}

func (dp *DummyPlugin) StartSequence() {}

func (dp *DummyPlugin) EndSequence() {
	eosUnit := common.MakeReqUnit(dp.name, common.EOS_REQUEST)
	common.Post_request(dp.callback, dp.name, eosUnit)
}

func (dp *DummyPlugin) DeliverUnit(unit common.CmUnit) {
	// Ensure correct order of calling by suspending worker thread
	if dp.role == 0 {
		time.Sleep(time.Second)
	}

	buf := 0
	if unit == nil {
		buf = 20
	} else {
		buf, _ = unit.GetBuf().(int)
	}

	dp.inCnt += buf

	if buf > 10 {
		reqUnit := common.MakeReqUnit(dp.name, common.FETCH_REQUEST)
		common.Post_request(dp.callback, dp.name, reqUnit)
	}

	if dp.role == 0 {
		eosUnit := common.MakeReqUnit(dp.name, common.EOS_REQUEST)
		common.Post_request(dp.callback, dp.name, eosUnit)
	}
}

func (dp *DummyPlugin) FetchUnit() common.CmUnit {
	rv := common.IOUnit{Buf: dp.inCnt*10 + dp.fetchCnt, IoType: 0, Id: -1}
	dp.fetchCnt += 1
	return rv
}

func (dp *DummyPlugin) SetCallback(callback common.RequestHandler) {
	dp.callback = callback
}

func GetDummyPlugin(name string, isRoot int) DummyPlugin {
	rv := DummyPlugin{logger: logs.CreateLogger("Dummy"), inCnt: 0, fetchCnt: 0, name: name, role: isRoot}
	rv.SetCallback(func(s string, reqType common.WORKER_REQUEST, obj interface{}) {})
	return rv
}
