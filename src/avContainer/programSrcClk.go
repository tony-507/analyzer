package avContainer

import (
	"fmt"
)

type programSrcClk struct {
	pcr       []int       // PCR value from input stream
	pcrLoc    []int       // Location of the PCR value
	curMaxLoc int         // Max index that contains PCR, not read from pcrLoc to prevent race condition
	streamCnt map[int]int // The packet count of each stream received, used for dropping outdated pcr values
	callback  *TsDemuxer  // Callback to demuxer
}

func getProgramSrcClk(callback *TsDemuxer) programSrcClk {
	return programSrcClk{pcr: make([]int, 0), pcrLoc: make([]int, 0), streamCnt: make(map[int]int, 0), callback: callback}
}

func (clk *programSrcClk) updatePcrRecord(pcr int, pktCnt int) {
	if pcr == -1 {
		// Does not carry a pcr
		return
	}
	// fmt.Println("Source time updated:", pcr, pktCnt)

	clk.pcr = append(clk.pcr, pcr)
	clk.pcrLoc = append(clk.pcrLoc, pktCnt)
	clk.curMaxLoc = pktCnt
}

// TODO How to ensure this is called after corresponding PCR is obtained?
func (clk *programSrcClk) requestPcr(pid int, curCnt int) (int, int) {
	if pid != -1 {
		// For removing old records
		clk.streamCnt[pid] = curCnt
	}

	// The first PCR record containing packet has payload within 1 packet
	if len(clk.pcrLoc) == 1 && clk.pcrLoc[0] == curCnt {
		return clk.pcr[0], 0
	}

	id0 := 0
	for i := 0; i < len(clk.pcr); i++ {
		if clk.pcrLoc[i] >= curCnt {
			id0 = i
			break
		}
	}

	if pid == -1 || (id0 == 0 && curCnt > clk.pcrLoc[0]) {
		// Out of PCR bound, do extrapolation by assuming CBR and raise an alarm to indicate possibility of PCR error
		last := len(clk.pcr) - 1
		step188 := (clk.pcr[last] - clk.pcr[last-1]) / (clk.pcrLoc[last] - clk.pcrLoc[last-1])
		curPcr := clk.pcr[last] + (curCnt-clk.pcrLoc[last])*step188
		if pid != -1 {
			infoMsg := fmt.Sprintf("Extrapolation of PCR is done at pkt#%d. Max PCR location found is %d", curCnt, clk.pcrLoc[len(clk.pcrLoc)-1])
			clk.callback.sendStatus(ctrl_INFO, pid, ctrl_NUMERICAL_RISK, curCnt, infoMsg)
		}
		return curPcr, 0
	} else if id0 == 0 {
		// Before first PCR record
		return -1, 1
	}

	curPcr := int((clk.pcr[id0]-clk.pcr[id0-1])*(curCnt-clk.pcrLoc[id0-1])/(clk.pcrLoc[id0]-clk.pcrLoc[id0-1])) + clk.pcr[id0-1]
	return curPcr, 0
}

func (clk *programSrcClk) _removeOldRecords() {
	min := int(^int(0) >> 1)
	idxToKeep := 0
	for _, cnt := range clk.streamCnt {
		if cnt < min {
			min = cnt
		}
	}
	for idx, loc := range clk.pcrLoc {
		if loc > min {
			if idx != 0 {
				idxToKeep = idx - 1
			}
			break
		}
	}
	clk.pcr = clk.pcr[idxToKeep:]
	clk.pcrLoc = clk.pcrLoc[idxToKeep:]
}
