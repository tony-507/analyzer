package h264

import (
	"github.com/tony-507/analyzers/src/common"
	"github.com/tony-507/analyzers/src/common/io"
)

type PicTiming struct{
	PicStructPresentFlag bool
	Clocks []PicClock
}

type PicClock struct {
	CtType int
	CountingType int
	DiscontinuityFlag bool
	CntDroppedFlag bool
	Tc common.TimeCode
}

func ParsePicTiming(r *io.BsReader, sequenceParameterSet SequenceParameterSet) PicTiming {
	picTiming := PicTiming{}
	cpbDpbDelaysPresentFlag := false
	if sequenceParameterSet.Vui.BPresent {
		if sequenceParameterSet.Vui.NalHrdParametersPresentFlag {
			cpbDpbDelaysPresentFlag = true
		} else if sequenceParameterSet.Vui.VclHrdParametersPresentFlag {
			cpbDpbDelaysPresentFlag = true
		}
	}

	if cpbDpbDelaysPresentFlag {
		cpbRemovalDelayLength := 24
		dpbOutputDelayLength := 24
		if sequenceParameterSet.Vui.Hrd.BPresent {
			cpbRemovalDelayLength = sequenceParameterSet.Vui.Hrd.CpbRemovalDelayLengthMinus1 + 1
			dpbOutputDelayLength = sequenceParameterSet.Vui.Hrd.DpbOutputDelayLengthMinus1 + 1
		}
		r.ReadBits(cpbRemovalDelayLength) // cpb_removal_delay
		r.ReadBits(dpbOutputDelayLength) // dpb_output_delay
	}

	picTiming.PicStructPresentFlag = sequenceParameterSet.Vui.PicStructPresentFlag

	if picTiming.PicStructPresentFlag {
		picStruct := r.ReadBits(4)
		numClockTs := 0
		switch picStruct {
		case 0:
			numClockTs = 1
		case 1:
			numClockTs = 1
		case 2:
			numClockTs = 1
		case 3:
			numClockTs = 2
		case 4:
			numClockTs = 2
		case 5:
			numClockTs = 3
		case 6:
			numClockTs = 3
		case 7:
			numClockTs = 2
		case 8:
			numClockTs = 3
		}
		for i := 0; i < numClockTs; i++ {
			clockTimestampFlag := r.ReadBits(1) != 0
			if clockTimestampFlag {
				picClock := PicClock{}
				picClock.CtType = r.ReadBits(2)
				r.ReadBits(1) // nuit_field_based_flag

				picClock.CountingType = r.ReadBits(5)
				fullTimestampFlag := r.ReadBits(1) != 0
				picClock.DiscontinuityFlag = r.ReadBits(1) != 0
				picClock.CntDroppedFlag = r.ReadBits(1) != 0
				nFrames := r.ReadBits(8)
				seconds := -1
				minutes := -1
				hours := -1
				if fullTimestampFlag {
					seconds = r.ReadBits(6)
					minutes = r.ReadBits(6)
					hours = r.ReadBits(5)
				} else {
					if r.ReadBits(1) != 0 {
						seconds = r.ReadBits(6)
						if r.ReadBits(1) != 0 {
							minutes = r.ReadBits(6)
							if r.ReadBits(1) != 0 {
								hours = r.ReadBits(5)
							}
						}
					}
				}
				picClock.Tc = common.TimeCode{
					Hour: hours,
					Minute: minutes,
					Second: seconds,
					Frame: nFrames,
					DropFrame: picClock.CountingType == 4,
				}

				if sequenceParameterSet.Vui.Hrd.TimeOffsetLength != 0 {
					r.ReadBits(sequenceParameterSet.Vui.Hrd.TimeOffsetLength)
				}

				picTiming.Clocks = append(picTiming.Clocks, picClock)
			}
		}
	}
	return picTiming
}
