package h264

type SequenceParameterSet struct {
	Vui VuiParameters
}

type VuiParameters struct {
	BPresent                    bool
	NalHrdParametersPresentFlag bool
	VclHrdParametersPresentFlag bool
	Hrd                         HrdParameters
	PicStructPresentFlag        bool
}

type HrdParameters struct {
	BPresent                    bool
	CpbRemovalDelayLengthMinus1 int
	DpbOutputDelayLengthMinus1  int
	TimeOffsetLength            int
}

func ParseSequenceParameterSet(rbsp []byte) SequenceParameterSet {
	sqp := CreateSequenceParameterSet()
	sqp.Vui.PicStructPresentFlag = true

	return sqp
}

func CreateSequenceParameterSet() SequenceParameterSet {
	return SequenceParameterSet{
		Vui: VuiParameters{
			PicStructPresentFlag: true,
			Hrd: HrdParameters{
				TimeOffsetLength: 24,
			},
		},
	}
}
