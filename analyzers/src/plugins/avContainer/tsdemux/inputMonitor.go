package tsdemux

import (
	"github.com/tony-507/analyzers/src/common"
	"github.com/tony-507/analyzers/src/plugins/avContainer/model"
)

const (
	_CC_ERROR int = 0
)

// A struct that monitors the input source
// It looks for error in headers
type inputMonitor struct {
	logger common.Log
	ccMap  map[int]int // pid -> cc
}

func (tm *inputMonitor) checkTsHeader(th model.TsHeader, pktCnt int) {
	// Look for CC error
	if currCC, hasKey := tm.ccMap[th.Pid]; hasKey {
		hasCcError := (th.Afc != 2 && th.Cc != (currCC+1)%16) || (th.Afc == 2 && th.Cc != currCC)
		hasCcError = th.Pid != 8191 && hasCcError
		if hasCcError {
			tm.logger.Error("CC error for pid %d at pkt#%d. Expected %d, but got %d", th.Pid, pktCnt, (currCC+1)%16, th.Cc)
		}
	}
	tm.ccMap[th.Pid] = th.Cc
}

func setupInputMonitor() inputMonitor {
	tsMon := inputMonitor{ccMap: map[int]int{}, logger: common.CreateLogger("inputMonitor")}

	return tsMon
}

var inputMon inputMonitor = setupInputMonitor()
