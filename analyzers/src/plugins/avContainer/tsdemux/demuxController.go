package tsdemux

import (
	"fmt"
	"strings"
	"sync"

	"github.com/tony-507/analyzers/src/tttKernel"
)

// Stream type conversion helper
type PKT_TYPE string

const (
	UNKNOWN PKT_TYPE = "undefined"
	PAT     PKT_TYPE = "PAT"
	PMT     PKT_TYPE = "PMT"
	SDT     PKT_TYPE = "SDT"
	VIDEO   PKT_TYPE = "video"
	AUDIO   PKT_TYPE = "audio"
	DATA    PKT_TYPE = "data"
)

// Internal controller for demuxer
// This enhances data protection among structs and provides higher flexibility
type demuxController struct {
	isRunning      bool                   // State of demuxer
	inputCnt       int
	outputQueueLen int
	progClkMap     map[int]*programSrcClk // progNum -> srcClk
	pktCntMap      map[int]int            // pid -> # of packets
	resourceLoader *tttKernel.ResourceLoader
	mtx            sync.Mutex
}

func (dc *demuxController) printInfo(sb *strings.Builder) {
	dc.mtx.Lock()
	sb.WriteString(fmt.Sprintf("\tCurrent count: %d\n", dc.inputCnt))
	sb.WriteString(fmt.Sprintf("\tisRunning: %v\n", dc.isRunning))
	sb.WriteString(fmt.Sprintf("\tOutput queue length: %d\n", dc.outputQueueLen))
	sb.WriteString("\tPacket statistics map:\n")
	for pid, cnt := range dc.pktCntMap {
		sb.WriteString(fmt.Sprintf("\t\t%3d: %7d\n", pid, cnt))
	}
	dc.mtx.Unlock()
}

func (dc *demuxController) inputReceived() {
	dc.mtx.Lock()
	dc.inputCnt += 1
	dc.mtx.Unlock()
}

func (dc *demuxController) getInputCount() int {
	return dc.inputCnt
}

func (dc *demuxController) dataParsed(pid int) {
	dc.mtx.Lock()
	if _, hasPid := dc.pktCntMap[pid]; !hasPid {
		dc.pktCntMap[pid] = 0
	}
	dc.pktCntMap[pid] += 1
	dc.mtx.Unlock()
}

func (dc *demuxController) outputUnitAdded() {
	dc.outputQueueLen += 1
}

func (dc *demuxController) updateSrcClk(progNum int) *programSrcClk {
	_, hasKey := dc.progClkMap[progNum]
	if !hasKey {
		clk := getProgramSrcClk(dc)
		dc.progClkMap[progNum] = &clk
	}
	return dc.progClkMap[progNum]
}

func (dc *demuxController) setResource(resourceLoader *tttKernel.ResourceLoader) {
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
	dc.outputQueueLen -= 1
}

func getControl() *demuxController {
	rv := demuxController{
		isRunning:  true,
		inputCnt:      0,
		outputQueueLen:     0,
		progClkMap: make(map[int]*programSrcClk, 0),
		pktCntMap:  make(map[int]int, 0),
	}
	return &rv
}
