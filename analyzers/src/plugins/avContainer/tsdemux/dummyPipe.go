package tsdemux

import (
	"github.com/tony-507/analyzers/src/common"
)

type dummyPipe struct {
	callback    *tsDemuxerPlugin
	outputQueue []common.CmUnit
	ready       bool
	inCnt       int
}

func (dp *dummyPipe) processUnit(buf []byte, pktCnt int) {
	// Return dummy unit
	dp.inCnt += 1
	if dp.inCnt > 1 {
		dp.ready = true
	}
	dummy := common.MakeIOUnit(buf, 3, -1)
	if dp.ready {
		dp.outputQueue = append(dp.outputQueue, dummy)
	}
}

func (dp *dummyPipe) getDuration() int {
	return 1
}

func (dp *dummyPipe) getProgramNumber(idx int) int {
	return 1
}

func (dp *dummyPipe) readyForFetch() bool {
	return dp.ready
}

func (dp *dummyPipe) getOutputUnit() common.CmUnit {
	if len(dp.outputQueue) == 0 {
		panic("[DummyPipe] Fatal error: Fetching from an empty output queue")
	}
	outUnit := dp.outputQueue[0]
	if len(dp.outputQueue) == 1 {
		dp.outputQueue = make([]common.CmUnit, 0)
	} else {
		dp.outputQueue = dp.outputQueue[1:]
	}
	return outUnit
}

func getDummyPipe(callback *tsDemuxerPlugin) dummyPipe {
	return dummyPipe{callback: callback, ready: false, inCnt: 0}
}
