package audio

import (
	"fmt"

	"github.com/tony-507/analyzers/src/common"
	"github.com/tony-507/analyzers/src/plugins/dataHandler/utils"
)

type ac3Handler struct {
	pid        int
	pesCnt     int
	curPesSize int
	bInit      bool
}

func (h *ac3Handler) Feed(unit common.CmUnit, newData *utils.ParsedData) error {
	h.pesCnt += 1
	cmBuf := unit.GetBuf()
	size, ok := common.GetBufFieldAsInt(cmBuf, "size")
	if !ok {
		return nil
	}
	if h.curPesSize != size {
		if h.curPesSize == 0 {
			h.curPesSize = size
		}
		if h.curPesSize != size {
			fmt.Println(fmt.Sprintf("[%d] At PES packet #%d, PES packet size changes from %d to %d", h.pid, h.pesCnt, h.curPesSize, size))
			h.curPesSize = size
		}
	}
	return nil
}

func AC3Handler(pid int) utils.DataHandler {
	return &ac3Handler{pid: pid, pesCnt: 0, curPesSize: 0, bInit: false}
}
