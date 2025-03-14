package tsdemux

import (
	"github.com/tony-507/analyzers/src/plugins/common"
	"github.com/tony-507/analyzers/src/tttKernel"
)

type dummyPipe struct {
	callback    IDemuxCallback
	outputQueue []tttKernel.CmUnit
	ready       bool
	inCnt       int
}

func (dp *dummyPipe) processUnit(buf []byte, pktCnt int) error {
	// Return dummy unit
	dp.inCnt += 1
	if dp.inCnt > 1 {
		dp.ready = true
	}
	dummy := common.NewMediaUnit(tttKernel.MakeSimpleBuf(buf), common.UNKNOWN_UNIT)
	if dp.ready {
		dp.outputQueue = append(dp.outputQueue, dummy)
		dp.callback.outputReady()
	}
	return nil
}

func (dp *dummyPipe) getDuration() int {
	return 1
}

func (dp *dummyPipe) getOutputUnit() tttKernel.CmUnit {
	if len(dp.outputQueue) == 0 {
		panic("[DummyPipe] Fatal error: Fetching from an empty output queue")
	}
	outUnit := dp.outputQueue[0]
	if len(dp.outputQueue) == 1 {
		dp.outputQueue = make([]tttKernel.CmUnit, 0)
	} else {
		dp.outputQueue = dp.outputQueue[1:]
	}
	return outUnit
}

func (dp *dummyPipe) start() {}

func (dp *dummyPipe) stop() {}

func getDummyPipe(callback IDemuxCallback) dummyPipe {
	return dummyPipe{
		callback: callback,
		ready: false,
		inCnt: 0,
	}
}
