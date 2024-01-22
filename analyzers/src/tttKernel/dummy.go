package tttKernel

import (
	"strings"

	"github.com/tony-507/analyzers/src/logging"
)

// This file stores some dummy struct for testing

type dummyUnit struct {
	buf CmBuf
}

func (u *dummyUnit) GetBuf() CmBuf {
	return u.buf
}

func (u *dummyUnit) GetField(name string) interface{} {
	return nil
}

type dummyPlugin struct {
	logger   logging.Log
	inCnt    int
	fetchCnt int
	callback RequestHandler
	name     string
	role     int // 0 represents a root, 1 represents non-root
}

func (dp *dummyPlugin) SetParameter(m_parameter string) {}

func (dp *dummyPlugin) SetResource(resourceLoader *ResourceLoader) {}

func (dp *dummyPlugin) StartSequence() {
	if dp.role == 0 {
		go dp.DeliverUnit(nil, "")
	}
}

func (dp *dummyPlugin) EndSequence() {
	dp.logger.Info("endSequence called")
	if dp.role != 0 {
		eosUnit := MakeReqUnit(dp.name, EOS_REQUEST)
		Post_request(dp.callback, dp.name, eosUnit)
	}
}

func (dp *dummyPlugin) DeliverUnit(unit CmUnit, inputId string) {
	// Ensure correct order of calling by suspending worker thread

	buf := 0
	if unit == nil {
		buf = 20
	} else {
		buf = int(GetBytesInBuf(unit)[0])
	}
	dp.logger.Info("deliverUnit called with buffer %d", buf)

	dp.inCnt += buf

	if buf > 10 {
		reqUnit := MakeReqUnit(dp.name, FETCH_REQUEST)
		Post_request(dp.callback, dp.name, reqUnit)
	}

	if dp.role == 0 {
		eosUnit := MakeReqUnit(dp.name, EOS_REQUEST)
		Post_request(dp.callback, dp.name, eosUnit)
	}
}

func (dp *dummyPlugin) DeliverStatus(unit CmUnit) {}

func (dp *dummyPlugin) FetchUnit() CmUnit {
	val := dp.inCnt * 10 + dp.fetchCnt
	rv := &dummyUnit{buf: MakeSimpleBuf([]byte{byte(val)})}
	dp.logger.Info("fetchUnit called with data %d", val)
	dp.fetchCnt += 1
	return rv
}

func (dp *dummyPlugin) SetCallback(callback RequestHandler) {
	dp.callback = callback
}

func (dp *dummyPlugin) PrintInfo(sb *strings.Builder) {}

func (dp *dummyPlugin) Name() string {
	return dp.name
}

func dummy(name string, role int) IPlugin {
	rv := dummyPlugin{name: name, logger: logging.CreateLogger(name), role: role}
	return &rv
}
