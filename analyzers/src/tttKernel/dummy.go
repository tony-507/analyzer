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

func (dp *DummyPlugin) setParameter(m_parameter string) {
	dp.logger.Info("setParameter called")
}

func (dp *DummyPlugin) setResource(resourceLoader *common.ResourceLoader) {
	dp.logger.Info("setResource called")
}

func (dp *DummyPlugin) startSequence() {
	dp.logger.Info("startSequence called")
}

func (dp *DummyPlugin) endSequence() {
	dp.logger.Info("endSequence called")
	eosUnit := common.MakeReqUnit(dp.name, common.EOS_REQUEST)
	common.Post_request(dp.callback, dp.name, eosUnit)
}

func (dp *DummyPlugin) deliverUnit(unit common.CmUnit) {
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

func (dp *DummyPlugin) deliverStatus(unit common.CmUnit) {
	dp.logger.Info("deliverStatus called with status %v", unit)
}

func (dp *DummyPlugin) fetchUnit() common.CmUnit {
	dp.logger.Info("fetchUnit called")
	rv := common.MakeIOUnit(dp.inCnt*10+dp.fetchCnt, 0, -1)
	dp.fetchCnt += 1
	return rv
}

func (dp *DummyPlugin) setCallback(callback common.RequestHandler) {
	dp.logger.Info("setCallback called")
	dp.callback = callback
}

func getDummyPlugin(name string, isRoot int) common.Plugin {
	rv := DummyPlugin{name: name, logger: common.CreateLogger(name), role: isRoot}
	return common.CreatePlugin(
		name,
		isRoot == 0,
		rv.setCallback,
		rv.setParameter,
		rv.setResource,
		rv.startSequence,
		rv.deliverUnit,
		rv.deliverStatus,
		rv.fetchUnit,
		rv.endSequence,
	)
}
