package tsdemux

import (
	"github.com/tony-507/analyzers/src/common"
)

type dummyPipe struct {
	callback    IDemuxCallback
	outputQueue []common.CmUnit
	ready       bool
	inCnt       int
}

func (dp *dummyPipe) processUnit(buf []byte, pktCnt int) error {
	// Return dummy unit
	dp.inCnt += 1
	if dp.inCnt > 1 {
		dp.ready = true
	}
	dummy := common.NewMediaUnit(common.MakeSimpleBuf(buf), common.UNKNOWN_UNIT)
	if dp.ready {
		dp.outputQueue = append(dp.outputQueue, dummy)
		dp.callback.outputReady()
	}
	return nil
}

func (dp *dummyPipe) getDuration() int {
	return 1
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

func getDummyPipe(callback IDemuxCallback) dummyPipe {
	return dummyPipe{
		callback: callback,
		ready: false,
		inCnt: 0,
	}
}
