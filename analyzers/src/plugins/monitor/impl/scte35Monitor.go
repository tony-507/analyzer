package impl

import (
	"fmt"

	"github.com/tony-507/analyzers/src/logging"
	"github.com/tony-507/analyzers/src/plugins/common"
	"github.com/tony-507/analyzers/src/tttKernel"
)

/*
 * This monitor displays video PTS, SCTE-35 splice time and SCTE-35 pre-roll time.
 *
 * It can be used to validate IDR insertion correctness and splice repetition.
 */

type scte35Data struct {
	pid int
	pts int64
	spliceTime int64
	preRoll int64
}

type Scte35Monitor struct {
	logger logging.Log
	inputIds []string
	data [][]scte35Data // inputId -> scte35Data
}

func (m *Scte35Monitor) Feed(unit tttKernel.CmUnit, inputId string) {
	if !m.HasInputId(inputId) {
		m.inputIds = append(m.inputIds, inputId)
		m.data = append(m.data, make([]scte35Data, 0))
	}
	mUnit, ok := unit.(*common.MediaUnit)
	if !ok {
		return
	}
	if mUnit.GetType() != common.DATA_UNIT {
		return
	}
	data := mUnit.Data
	pts := data.GetField("playtime")
	spliceTime := data.GetField("spliceTime")
	preRoll := data.GetField("preroll")

	pid, _ := tttKernel.GetBufFieldAsInt(unit.GetBuf(), "pid")

	idx := m.getIdIndex(inputId)
	m.data[idx] = append(m.data[idx], scte35Data{pid, pts, spliceTime, preRoll / 90})
}

func (m *Scte35Monitor) GetFields() []string {
	return []string{"Pid", "PTS", "Splice Time", "Pre-roll"}
}

func (m *Scte35Monitor) HasInputId(inputId string) bool {
	return m.getIdIndex(inputId) != -1
}

func (m *Scte35Monitor) GetDisplayData() []string {
	maxLength := 0
	for _, v := range m.data {
		if maxLength < len(v) {
			maxLength = len(v)
		}
	}

	res := make([]string, maxLength)

	for id := range m.data {
		for idx := 0; idx < maxLength; idx++ {
			if idx < len(m.data[id]) {
				datum := m.data[id][idx]
				res[idx] += fmt.Sprintf("|%15d|%15d|%15d|%15d", datum.pid, datum.pts, datum.spliceTime, datum.preRoll)
			} else {
				res[idx] += fmt.Sprintf("|%15s|%15s|%15s|%15s", "", "", "", "")
			}
		}
	}

	for i := range res {
		res[i] += "|"
	}

	return res
}

func (m *Scte35Monitor) getIdIndex(inputId string) int {
	for i, id := range m.inputIds {
		if id == inputId {
			return i
		}
	}
	return -1
}

func GetScte35Monitor() MonitorImpl {
	return &Scte35Monitor{
		logger: logging.CreateLogger("Scte35Monitor"),
		inputIds: make([]string, 0),
		data: make([][]scte35Data, 0),
	}
}
