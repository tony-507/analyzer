package model

import (
	"github.com/tony-507/analyzers/src/common"
	"github.com/tony-507/analyzers/src/common/io"
)

var logger common.Log

type Splice_command interface {
	GetSplicePTS() []int
}

// SCTE-35 2019a section 9.7.1
type Splice_null struct{}

func (cmd Splice_null) GetSplicePTS() []int {
	return make([]int, 0)
}

// SCTE-35 2019a section 9.7.2
type Splice_schedule struct {
	SpliceCnt    int
	SpliceEvents []Splice_event
}

func (cmd Splice_schedule) GetSplicePTS() []int {
	// TODO: This is UTC splice time
	rv := []int{}
	for _, e := range cmd.SpliceEvents {
		rv = append(rv, e.GetSplicePTS()...)
	}
	return rv
}

type Splice_event struct {
	EventId             int
	EventCancelIdr      bool
	OutOfNetworkIdr     bool
	ProgramSpliceFlag   bool
	DurationFlag        bool
	SpliceImmediateFlag bool
	SpliceTime          int
	Components          []Splice_component
	BreakDuration       Break_duration
	UniqueProgramId     int
	AvailNum            int
	AvailsExpected      int
}

func (cmd Splice_event) GetSplicePTS() []int {
	if cmd.SpliceTime != -1 {
		return []int{cmd.SpliceTime}
	} else {
		rv := []int{}
		for _, component := range cmd.Components {
			rv = append(rv, component.SpliceTime)
		}
		return rv
	}
}

type Splice_component struct {
	ComponentTag int
	SpliceTime   int
}

// SCTE-35 2019a section 9.7.4
type Time_signal struct {
	SpliceTime int
}

func (cmd Time_signal) GetSplicePTS() []int {
	return []int{cmd.SpliceTime}
}

// SCTE-35 2019a section 9.7.6
type Private_command struct {
	Identifier   string
	PrivateBytes string
}

func (cmd Private_command) GetSplicePTS() []int {
	return make([]int, 0)
}

type Break_duration struct {
	AutoReturn bool
	Duration   int
}

func readSpliceSchedule(r *io.BsReader) Splice_schedule {
	spliceCnt := (*r).ReadBits(8)
	spliceEvents := []Splice_event{}

	for i := 0; i < spliceCnt; i++ {
		event := readSpliceEvent(r, false)
		spliceEvents = append(spliceEvents, event)
	}

	return Splice_schedule{SpliceCnt: spliceCnt, SpliceEvents: spliceEvents}
}

func readSpliceEvent(r *io.BsReader, isSpliceInsert bool) Splice_event {
	spliceEventId := (*r).ReadBits(32)
	spliceEventCancelIdr := (*r).ReadBits(1) != 0

	(*r).ReadBits(7)

	if !spliceEventCancelIdr {
		outOfNetworkIdr := (*r).ReadBits(1) != 0
		programSpliceFlag := (*r).ReadBits(1) != 0
		durationFlag := (*r).ReadBits(1) != 0
		spliceImmediateFlag := (*r).ReadBits(1) != 0

		(*r).ReadBits(4)

		spliceTime := -1
		components := []Splice_component{}

		bOnlyTime := false
		if isSpliceInsert {
			bOnlyTime = programSpliceFlag && !spliceImmediateFlag
		} else {
			bOnlyTime = programSpliceFlag
		}

		if bOnlyTime {
			spliceTime = readSpliceTime(r, isSpliceInsert)
		} else {
			componentCnt := (*r).ReadBits(8)
			for i := 0; i < componentCnt; i++ {
				tag := (*r).ReadBits(8)
				sTime := -1
				if !isSpliceInsert || (isSpliceInsert && !spliceImmediateFlag) {
					sTime = readSpliceTime(r, isSpliceInsert)
				}
				components = append(components, Splice_component{ComponentTag: tag, SpliceTime: sTime})
			}
		}

		breakDuration := Break_duration{}
		if durationFlag {
			breakDuration = readBreakDuration(r)
		}

		uniqueProgramId := (*r).ReadBits(16)
		availNum := (*r).ReadBits(8)
		availsExpected := (*r).ReadBits(8)

		return Splice_event{EventId: spliceEventId, EventCancelIdr: spliceEventCancelIdr, OutOfNetworkIdr: outOfNetworkIdr,
			ProgramSpliceFlag: programSpliceFlag, DurationFlag: durationFlag, SpliceImmediateFlag: spliceImmediateFlag,
			SpliceTime: spliceTime, Components: components, BreakDuration: breakDuration, UniqueProgramId: uniqueProgramId,
			AvailNum: availNum, AvailsExpected: availsExpected}
	} else {
		return Splice_event{EventId: spliceEventId, EventCancelIdr: spliceEventCancelIdr}
	}
}

func readTimeSignal(r *io.BsReader) Time_signal {
	spliceTime := readSpliceTime(r, true)
	return Time_signal{SpliceTime: spliceTime}
}

func readPrivateCommand(r *io.BsReader) Private_command {
	identifier := (*r).ReadChar(32)

	return Private_command{Identifier: identifier, PrivateBytes: (*r).ReadHex(len((*r).GetRemainedBuffer()))}
}

func readSpliceTime(r *io.BsReader, isSpliceTime bool) int {
	if isSpliceTime {
		flag := (*r).ReadBits(1) != 0
		if flag {
			(*r).ReadBits(6)
			return (*r).ReadBits(33)
		}
		(*r).ReadBits(7)
	} else {
		// UTC splice time
		return (*r).ReadBits(32)
	}
	return -1
}

func readBreakDuration(r *io.BsReader) Break_duration {
	autoReturn := (*r).ReadBits(1) != 0
	(*r).ReadBits(6)
	duration := (*r).ReadBits(33)
	return Break_duration{AutoReturn: autoReturn, Duration: duration}
}
