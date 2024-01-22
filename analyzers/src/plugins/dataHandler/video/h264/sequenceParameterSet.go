package h264

import (
	"github.com/tony-507/analyzers/src/plugins/common/io"
)

type SequenceParameterSet struct {
	Id      int
	Level   int
	Profile int
	Vui     VuiParameters
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
	r := io.GetBufferReader(rbsp)

	sqp.Profile = r.ReadBits(8)
	r.ReadBits(1) // constraint_set0_flag
	r.ReadBits(1) // constraint_set1_flag
	r.ReadBits(1) // constraint_set2_flag
	r.ReadBits(1) // constraint_set3_flag
	r.ReadBits(1) // constraint_set4_flag
	r.ReadBits(1) // constraint_set5_flag
	r.ReadAndAssertBits(2, 0, "Reserved bits not zero in sequence parameter set")
	sqp.Level = r.ReadBits(8)
	sqp.Id = r.ReadExpGolomb()

	if sqp.Profile == 100 || sqp.Profile == 110 || sqp.Profile == 122 || sqp.Profile == 244 ||
		sqp.Profile == 44 || sqp.Profile == 83 || sqp.Profile == 86 || sqp.Profile == 118 ||
		sqp.Profile == 128 {
		r.ReadExpGolomb() // chroma_format_idc
	}

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
