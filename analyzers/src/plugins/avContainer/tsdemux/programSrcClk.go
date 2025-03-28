package tsdemux

type programSrcClk struct {
	pcr       []int            // PCR value from input stream
	pcrLoc    []int            // Location of the PCR value
	curMaxLoc int              // Max index that contains PCR, not read from pcrLoc to prevent race condition
	eptStart  bool             // Indicate if extrapolation has started
	streamCnt map[int]int      // The packet count of each stream received, used for dropping outdated pcr values
	callback  *demuxController // Callback to demuxer
}

func getProgramSrcClk(callback *demuxController) programSrcClk {
	return programSrcClk{pcr: make([]int, 0), pcrLoc: make([]int, 0), streamCnt: make(map[int]int, 0), callback: callback}
}

func (clk *programSrcClk) updatePcrRecord(pcr int, pktCnt int) {
	if pcr == -1 {
		// Does not carry a pcr
		return
	}

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

	// Before first PCR record
	if len(clk.pcrLoc) < 2 {
		return -1, 1
	}

	id0 := 0
	for i := 0; i < len(clk.pcr); i++ {
		if clk.pcrLoc[i] >= curCnt {
			id0 = i
			break
		}
	}

	if pid == -1 || (id0 == 0 && curCnt > clk.pcrLoc[0]) {
		// Out of PCR bound, do extrapolation by assuming CBR
		last := len(clk.pcr) - 1
		step188 := (clk.pcr[last] - clk.pcr[last-1]) / (clk.pcrLoc[last] - clk.pcrLoc[last-1])
		curPcr := clk.pcr[last] + (curCnt-clk.pcrLoc[last])*step188
		if pid != -1 {
			if !clk.eptStart {
				clk.eptStart = true
			}
		}
		return curPcr, 0
	} else if id0 == 0 {
		// Not ready
		return -1, 1
	}

	curPcr := int((clk.pcr[id0]-clk.pcr[id0-1])*(curCnt-clk.pcrLoc[id0-1])/(clk.pcrLoc[id0]-clk.pcrLoc[id0-1])) + clk.pcr[id0-1]
	clk.eptStart = false
	return curPcr, 0
}
