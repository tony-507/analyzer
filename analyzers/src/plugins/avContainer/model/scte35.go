package model

import (
	"fmt"

	"github.com/tony-507/analyzers/src/common"
)

var logger common.Log

type Splice_command interface {
	GetSplicePTS() []int
}

// SCTE-35 2019a section 9.9
type Splice_info_section struct {
	TableId          int
	SectionSyntaxIdr bool
	PrivateIdr       bool
	SectionLen       int
	ProtocolVersion  int
	EncryptedPkt     bool
	EncryptAlgo      int
	PtsAdjustment    int
	CwIdx            int
	Tier             int
	SpliceCmdLen     int
	SpliceCmdType    int
	SpliceCmdTypeStr string
	SpliceCmd        Splice_command
	Crc32            int
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

// Truncated from the actual parsing function
func SCTE35ReadyForParse(buf []byte, afc int) bool {
	if afc > 1 {
		af := ParseAdaptationField(buf)
		buf = buf[(af.AfLen + 1):]
	}

	r := common.GetBufferReader(buf)

	pFieldLen := r.ReadBits(8)
	r.ReadBits(pFieldLen * 8)

	r.ReadBits(12)

	sectionLen := r.ReadBits(12)

	return r.GetSize() <= (sectionLen + 4 + pFieldLen)
}

func ReadSCTE35Section(buf []byte, afc int) Splice_info_section {
	logger = common.CreateLogger("SCTE35Parser")

	if afc > 1 {
		af := ParseAdaptationField(buf)
		buf = buf[(af.AfLen + 1):]
	}

	r := common.GetBufferReader(buf)

	pFieldLen := r.ReadBits(8)
	r.ReadBits(pFieldLen * 8)

	tableId := r.ReadBits(8)
	sectionSyntaxIdr := r.ReadBits(1) != 0
	privateIdr := r.ReadBits(1) != 0

	r.ReadBits(2)

	sectionLen := r.ReadBits(12)
	protocolVersion := r.ReadBits(8)
	encryptedPkt := r.ReadBits(1) != 0
	encryptedAlgo := r.ReadBits(6)
	ptsAdjustment := r.ReadBits(33)
	cwIdx := r.ReadBits(8)
	tier := r.ReadBits(12)

	spliceCmdLen := r.ReadBits(12)
	spliceCmdType := r.ReadBits(8)

	spliceCmdTypeStr := "Unknown"
	var spliceCmd Splice_command

	switch spliceCmdType {
	case 0x00:
		// Splice null, do nothing
		spliceCmdTypeStr = "splice_null"
		spliceCmd = Splice_null{}
	case 0x04:
		// Splice schedule
		spliceCmdTypeStr = "splice_schedule"
		spliceCmd = readSpliceSchedule(&r)
	case 0x05:
		// Splice insert
		spliceCmdTypeStr = "splice_insert"
		spliceCmd = readSpliceEvent(&r, true)
	case 0x06:
		// Time signal
		spliceCmdTypeStr = "time_signal"
		spliceCmd = readTimeSignal(&r)
	case 0x07:
		// Bandwidth reservation
		spliceCmdTypeStr = "bandwidth_reservation"
	case 0xff:
		// Private command
		spliceCmdTypeStr = "private_command"
		spliceCmd = readPrivateCommand(&r)
	default:
		msg := fmt.Sprint("Unknown splice command type ", spliceCmdType)
		logger.Error(msg)
		panic(msg)
	}

	return Splice_info_section{TableId: tableId, SectionSyntaxIdr: sectionSyntaxIdr, PrivateIdr: privateIdr,
		SectionLen: sectionLen, ProtocolVersion: protocolVersion, EncryptedPkt: encryptedPkt, EncryptAlgo: encryptedAlgo,
		PtsAdjustment: ptsAdjustment, CwIdx: cwIdx, Tier: tier, SpliceCmdLen: spliceCmdLen, SpliceCmdType: spliceCmdType, SpliceCmdTypeStr: spliceCmdTypeStr, SpliceCmd: spliceCmd}
}

func readSpliceSchedule(r *common.BsReader) Splice_schedule {
	spliceCnt := (*r).ReadBits(8)
	spliceEvents := []Splice_event{}

	for i := 0; i < spliceCnt; i++ {
		event := readSpliceEvent(r, false)
		spliceEvents = append(spliceEvents, event)
	}

	return Splice_schedule{SpliceCnt: spliceCnt, SpliceEvents: spliceEvents}
}

func readSpliceEvent(r *common.BsReader, isSpliceInsert bool) Splice_event {
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

func readTimeSignal(r *common.BsReader) Time_signal {
	spliceTime := readSpliceTime(r, true)
	return Time_signal{SpliceTime: spliceTime}
}

func readPrivateCommand(r *common.BsReader) Private_command {
	identifier := (*r).ReadChar(32)

	return Private_command{Identifier: identifier, PrivateBytes: (*r).ReadHex(len((*r).GetRemainedBuffer()))}
}

func readSpliceTime(r *common.BsReader, isSpliceTime bool) int {
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

func readBreakDuration(r *common.BsReader) Break_duration {
	autoReturn := (*r).ReadBits(1) != 0
	(*r).ReadBits(6)
	duration := (*r).ReadBits(33)
	return Break_duration{AutoReturn: autoReturn, Duration: duration}
}
