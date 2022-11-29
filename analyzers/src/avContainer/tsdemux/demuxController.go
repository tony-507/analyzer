package tsdemux

import (
	"fmt"
	"sync"
	"time"

	"github.com/tony-507/analyzers/src/resources"
)

// Internal controller for demuxer
// This enhances data protection among structs and provides higher flexibility
type demuxController struct {
	isRunning      bool                   // State of demuxer
	pollPeriod     int                    // Period for stuck detection
	inCnt          int                    // Current input count
	pCnt           int                    // Current parsing count
	outLen         int                    // Output queue length
	progClkMap     map[int]*programSrcClk // progNum -> srcClk
	pktCntMap      map[int]int            // pid -> # of packets
	resourceLoader *resources.ResourceLoader
	mtx            sync.Mutex
}

func (dc *demuxController) monitor() {
	for dc.isRunning {
		time.Sleep(time.Duration(dc.pollPeriod) * time.Second)
		dc.mtx.Lock()
		if dc.inCnt == dc.pCnt {
			statMsg := "tsDemuxer status\n"
			statMsg += fmt.Sprintf("\tCurrent count: %d\n", dc.inCnt)
			statMsg += fmt.Sprintf("\tisRunning: %v\n", dc.isRunning)
			statMsg += fmt.Sprintf("\tOutput queue size: %d\n", dc.outLen)
			fmt.Println(statMsg)
		}
		dc.mtx.Unlock()
	}
}

func (dc *demuxController) inputReceived() {
	dc.mtx.Lock()
	dc.inCnt += 1
	dc.mtx.Unlock()
}

func (dc *demuxController) getInputCount() int {
	return dc.inCnt
}

func (dc *demuxController) dataParsed(pid int) {
	dc.mtx.Lock()
	dc.pCnt += 1
	dc.mtx.Unlock()

	if _, hasPid := dc.pktCntMap[pid]; !hasPid {
		dc.pktCntMap[pid] = 0
	}
	dc.pktCntMap[pid] += 1
}

func (dc *demuxController) outputUnitAdded() {
	dc.outLen += 1
}

func (dc *demuxController) updateSrcClk(progNum int) *programSrcClk {
	_, hasKey := dc.progClkMap[progNum]
	if !hasKey {
		clk := getProgramSrcClk(dc)
		dc.progClkMap[progNum] = &clk
	}
	return dc.progClkMap[progNum]
}

func (dc *demuxController) setResource(resourceLoader *resources.ResourceLoader) {
	dc.resourceLoader = resourceLoader
}

func (dc *demuxController) queryStreamType(typeNum int) string {
	if dc.resourceLoader != nil {
		return dc.resourceLoader.Query("streamType", typeNum)
	} else {
		return "undefined"
	}
}

func (dc *demuxController) stop() {
	dc.isRunning = false
}

func (dc *demuxController) outputUnitFetched() {
	dc.outLen -= 1
}

func (dc *demuxController) throwError(pid int, curCnt int, msg string) {
	outMsg := fmt.Sprintf("[%d] At pkt#%d, %s", pid, curCnt, msg)
	panic(outMsg)
}

func (dc *demuxController) printSummary(duration int) {
	sum := 0
	rateSum := 0.0

	// TS statistics

	statMsg := "TS statistics:\n"
	statMsg += fmt.Sprintf("TS duration: %fs\n", float64(duration)/27000000)
	statMsg += "-------------------------------------------------\n"
	statMsg += "|    pid    |   count   |  bitrate  | frequency |\n"
	statMsg += "|-----------|-----------|-----------|-----------|\n"
	for pid, cnt := range dc.pktCntMap {
		rate := float64(cnt) * 1504 * 27000000 / float64(duration)
		rateSum += rate
		statMsg += fmt.Sprintf("|%11d|%11d|%11.2f|%11.2f|\n", pid, cnt, rate, rate/1504)
		sum += cnt
	}
	statMsg += "-------------------------------------------------\n"
	statMsg += fmt.Sprintf("|%11s|%11d|%11.2f|%11s|\n", "", sum, rateSum, "")
	statMsg += "-------------------------------------------------\n"

	fmt.Println(statMsg)
}

func getControl() *demuxController {
	rv := demuxController{isRunning: true, pollPeriod: 5, inCnt: 0, pCnt: 0, outLen: 0, progClkMap: make(map[int]*programSrcClk, 0), pktCntMap: make(map[int]int, 0)}
	return &rv
}
