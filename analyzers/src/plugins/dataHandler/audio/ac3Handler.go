package audio

import (
	"fmt"

	"github.com/tony-507/analyzers/src/tttKernel"
	"github.com/tony-507/analyzers/src/common/logging"
	"github.com/tony-507/analyzers/src/plugins/dataHandler/utils"
)

type ac3Handler struct {
	logger     logging.Log
	pid        int
	pesCnt     int
	curPesSize int
	bInit      bool
}

func (h *ac3Handler) Feed(unit tttKernel.CmUnit, newData *utils.ParsedData) error {
	h.pesCnt += 1
	cmBuf := unit.GetBuf()
	size, ok := tttKernel.GetBufFieldAsInt(cmBuf, "size")
	if !ok {
		return nil
	}
	if h.curPesSize != size {
		if h.curPesSize == 0 {
			h.curPesSize = size
		}
		if h.curPesSize != size {
			h.logger.Info("[%d] At PES packet #%d, PES packet size changes from %d to %d", h.pid, h.pesCnt, h.curPesSize, size)
			h.curPesSize = size
		}
	}
	return nil
}

func AC3Handler(pid int) utils.DataHandler {
	return &ac3Handler{
		logger: logging.CreateLogger(fmt.Sprintf("AC3_%d", pid)),
		pid: pid,
		pesCnt: 0,
		curPesSize: 0,
		bInit: false,
	}
}
