package tsdemux

import (
	"github.com/tony-507/analyzers/src/common"
)

type dummyPipe struct {
	callback *TsDemuxer
	ready    bool
	inCnt    int
}

func (dp *dummyPipe) processUnit(buf []byte, pktCnt int) {
	// Return dummy unit
	dp.inCnt += 1
	if dp.inCnt > 1 {
		dp.ready = true
	}
	dummy := common.IOUnit{Buf: buf, IoType: 3, Id: -1}
	if dp.ready {
		dp.callback.outputQueue = append(dp.callback.outputQueue, dummy)
	}
}

func (dp *dummyPipe) getDuration() int {
	return 1
}

func (dp *dummyPipe) getProgramNumber(idx int) int {
	return 1
}

func (dp *dummyPipe) clockReady() bool {
	return dp.ready
}

func getDummyPipe(callback *TsDemuxer) dummyPipe {
	return dummyPipe{callback: callback, ready: false, inCnt: 0}
}
