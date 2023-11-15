package tsdemux

import (
	"github.com/tony-507/analyzers/src/common/logging"
)

const (
	_CC_ERROR int = 0
)

// A struct that monitors the input source
// It looks for error in headers
type inputMonitor struct {
	logger logging.Log
	ccMap  map[int]int // pid -> cc
}

func (tm *inputMonitor) checkTsHeader(pid int, afc int, cc int, pktCnt int) {
	// Look for CC error
	if currCC, hasKey := tm.ccMap[pid]; hasKey {
		hasCcError := (afc != 2 && cc != (currCC+1)%16) || (afc == 2 && cc != currCC)
		hasCcError = pid != 8191 && hasCcError
		if hasCcError {
			tm.logger.Error("CC error for pid %d at pkt#%d. Expected %d, but got %d", pid, pktCnt, (currCC+1)%16, cc)
		}
	}
	tm.ccMap[pid] = cc
}

func setupInputMonitor() inputMonitor {
	tsMon := inputMonitor{ccMap: map[int]int{}, logger: logging.CreateLogger("inputMonitor")}

	return tsMon
}
