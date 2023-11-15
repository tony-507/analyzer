package tttKernel

import (
	"strings"

	"github.com/tony-507/analyzers/src/common"
	"github.com/tony-507/analyzers/src/common/logging"
)

// This file stores some dummy struct for testing
type dummyPlugin struct {
	logger   logging.Log
	inCnt    int
	fetchCnt int
	callback common.RequestHandler
	name     string
	role     int // 0 represents a root, 1 represents non-root
}

func (dp *dummyPlugin) SetParameter(m_parameter string) {}

func (dp *dummyPlugin) SetResource(resourceLoader *common.ResourceLoader) {}

func (dp *dummyPlugin) StartSequence() {
	if dp.role == 0 {
		go dp.DeliverUnit(nil, "")
	}
}

func (dp *dummyPlugin) EndSequence() {
	dp.logger.Info("endSequence called")
	if dp.role != 0 {
		eosUnit := common.MakeReqUnit(dp.name, common.EOS_REQUEST)
		common.Post_request(dp.callback, dp.name, eosUnit)
	}
}

func (dp *dummyPlugin) DeliverUnit(unit common.CmUnit, inputId string) {
	// Ensure correct order of calling by suspending worker thread

	buf := 0
	if unit == nil {
		buf = 20
	} else {
		buf = int(common.GetBytesInBuf(unit)[0])
	}
	dp.logger.Info("deliverUnit called with buffer %d", buf)

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

func (dp *dummyPlugin) DeliverStatus(unit common.CmUnit) {}

func (dp *dummyPlugin) FetchUnit() common.CmUnit {
	val := dp.inCnt * 10 + dp.fetchCnt
	rv := common.MakeIOUnit(common.MakeSimpleBuf([]byte{byte(val)}), 0, -1)
	dp.logger.Info("fetchUnit called with data %d", val)
	dp.fetchCnt += 1
	return rv
}

func (dp *dummyPlugin) SetCallback(callback common.RequestHandler) {
	dp.callback = callback
}

func (dp *dummyPlugin) PrintInfo(sb *strings.Builder) {}

func (dp *dummyPlugin) Name() string {
	return dp.name
}

func dummy(name string, role int) common.IPlugin {
	rv := dummyPlugin{name: name, logger: logging.CreateLogger(name), role: role}
	return &rv
}
