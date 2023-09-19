package tttKernel

import (
	"time"

	"github.com/tony-507/analyzers/src/common"
)

// This file stores some dummy struct for testing
type DummyPlugin struct {
	logger   common.Log
	inCnt    int
	fetchCnt int
	callback common.RequestHandler
	name     string
	role     int // 0 represents a root, 1 represents non-root
}

func (dp *DummyPlugin) SetParameter(m_parameter string) {
	dp.logger.Info("setParameter called")
}

func (dp *DummyPlugin) SetResource(resourceLoader *common.ResourceLoader) {
	dp.logger.Info("setResource called")
}

func (dp *DummyPlugin) StartSequence() {
	dp.logger.Info("startSequence called")
	if dp.role == 0 {
		go dp.DeliverUnit(nil)
	}
}

func (dp *DummyPlugin) EndSequence() {
	dp.logger.Info("endSequence called")
	eosUnit := common.MakeReqUnit(dp.name, common.EOS_REQUEST)
	common.Post_request(dp.callback, dp.name, eosUnit)
}

func (dp *DummyPlugin) DeliverUnit(unit common.CmUnit) {
	dp.logger.Info("deliverUnit called with unit %v", unit)
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

func (dp *DummyPlugin) DeliverStatus(unit common.CmUnit) {
	dp.logger.Info("deliverStatus called with status %v", unit)
}

func (dp *DummyPlugin) FetchUnit() common.CmUnit {
	rv := common.MakeIOUnit(dp.inCnt*10+dp.fetchCnt, 0, -1)
	dp.logger.Info("fetchUnit called with unit %v", rv)
	dp.fetchCnt += 1
	return rv
}

func (dp *DummyPlugin) SetCallback(callback common.RequestHandler) {
	dp.logger.Info("setCallback called")
	dp.callback = callback
}

func (dp *DummyPlugin) Name() string {
	return dp.name
}

func Dummy(name string, role int) common.IPlugin {
	rv := DummyPlugin{name: name, logger: common.CreateLogger(name), role: role}
	return &rv
}
