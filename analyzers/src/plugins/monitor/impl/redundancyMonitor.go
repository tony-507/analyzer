package impl

import (
	"fmt"

	"github.com/tony-507/analyzers/src/common"
)

var _MONITOR_QUEUE_SIZE = 10

type redundancyMonitor struct {
	dataQueues map[string][]common.CmUnit
	timeReference redundancyTimeRef
}

func (rm *redundancyMonitor) Feed(unit common.CmUnit, inputId string) {
	// Ensure input id is added to the map
	if !rm.HasInputId(inputId) {
		rm.dataQueues[inputId] = make([]common.CmUnit, _MONITOR_QUEUE_SIZE)
	}

	if rm.timeReference == Vitc {
		if _, hasTimeCode := common.GetBufFieldAsString(unit.GetBuf(), "timecode"); !hasTimeCode {
			return
		}
	}

	if len(rm.dataQueues[inputId]) == _MONITOR_QUEUE_SIZE {
		rm.dataQueues[inputId] = append(rm.dataQueues[inputId][1:], unit)
	} else {
		rm.dataQueues[inputId] = append(rm.dataQueues[inputId], unit)
	}
}

func (rm *redundancyMonitor) GetFields() []string {
	if rm.timeReference == Vitc {
		return []string{"PTS", "VITC"}
	} else {
		return []string{"PTS"}
	}
}

func (rm *redundancyMonitor) HasInputId(inputId string) bool {
	_, hasKey := rm.dataQueues[inputId]
	return hasKey
}

func (rm *redundancyMonitor) GetDisplayData() []string {
	l := 0
	for _, v := range rm.dataQueues {
		if l < len(v) {
			l = len(v)
		}
	}

	res := make([]string, l)

	for _, v := range rm.dataQueues {
		for idx, datum := range v {
			if datum == nil {
				continue
			}
			pts, _ := common.GetBufFieldAsInt(datum.GetBuf(), "pts")
			res[_MONITOR_QUEUE_SIZE - 1 - idx] += fmt.Sprintf("|%15d", pts)
			if rm.timeReference == Vitc {
				tc, _ := common.GetBufFieldAsString(datum.GetBuf(), "timecode")
				res[_MONITOR_QUEUE_SIZE - 1 - idx] += fmt.Sprintf("|%15s", tc)
			}
		}
	}

	for i := range res {
		res[i] += "|"
	}

	return res
}

func GetRedundancyMonitor(param *RedundancyParam) MonitorImpl {
	return &redundancyMonitor{
		dataQueues: map[string][]common.CmUnit{},
		timeReference: param.TimeRef,
	}
}
