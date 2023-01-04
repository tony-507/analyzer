package tttKernel

import (
	"fmt"
	"time"

	"github.com/tony-507/analyzers/src/common"
	"github.com/tony-507/analyzers/src/logs"
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

func (dp *DummyPlugin) setParameter(m_parameter string) {
	fmt.Println(fmt.Sprintf("[%s] setParameter called", dp.name))
}

func (dp *DummyPlugin) setResource(resourceLoader *common.ResourceLoader) {
	fmt.Println(fmt.Sprintf("[%s] setResource called", dp.name))
}

func (dp *DummyPlugin) startSequence() {
	fmt.Println(fmt.Sprintf("[%s] startSequence called", dp.name))
}

func (dp *DummyPlugin) endSequence() {
	fmt.Println(fmt.Sprintf("[%s] endSequence called", dp.name))
	eosUnit := common.MakeReqUnit(dp.name, common.EOS_REQUEST)
	common.Post_request(dp.callback, dp.name, eosUnit)
}

func (dp *DummyPlugin) deliverUnit(unit common.CmUnit) {
	fmt.Println(fmt.Sprintf("[%s] deliverUnit called with unit %v", dp.name, unit))
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
	fmt.Println(fmt.Sprintf("[%s] deliverStatus called with status %v", dp.name, unit))
}

func (dp *DummyPlugin) fetchUnit() common.CmUnit {
	fmt.Println(fmt.Sprintf("[%s] fetchUnit called", dp.name))
	rv := common.IOUnit{Buf: dp.inCnt*10 + dp.fetchCnt, IoType: 0, Id: -1}
	dp.fetchCnt += 1
	return rv
}

func (dp *DummyPlugin) setCallback(callback common.RequestHandler) {
	fmt.Println(fmt.Sprintf("[%s] setCallback called", dp.name))
	dp.callback = callback
}

func getDummyPlugin(name string, isRoot int) common.Plugin {
	rv := DummyPlugin{name: name, logger: logs.CreateLogger(name), role: isRoot}
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
