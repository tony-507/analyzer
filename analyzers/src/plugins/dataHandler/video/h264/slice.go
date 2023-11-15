package h264

import (
	"github.com/tony-507/analyzers/src/common/io"
	"github.com/tony-507/analyzers/src/plugins/dataHandler/utils"
)

func ReadSlice(r *io.BsReader, data *utils.VideoDataStruct) {
	readSliceHeader(r, data)
}

func readSliceHeader(r *io.BsReader, data *utils.VideoDataStruct) {
	r.ReadExpGolomb() // first_mb_in_slice
	slice_type := r.ReadExpGolomb()

	switch slice_type % 5 {
	case 0:
		data.Type = utils.P_SLICE
	case 1:
		data.Type = utils.B_SLICE
	case 2:
		data.Type = utils.I_SLICE
	default:
		data.Type = utils.UNKNOWN_SLICE
	}

	r.ReadBits((len(r.GetRemainedBuffer()) - 1) * 8 + r.GetOffset())
}