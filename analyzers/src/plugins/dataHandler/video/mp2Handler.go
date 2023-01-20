package video

import (
	"fmt"

	"github.com/tony-507/analyzers/src/common"
	"github.com/tony-507/analyzers/src/plugins/dataHandler/utils"
)

type _MP2_VIDEO_START_CODE int

const (
	_PICTURE_START     _MP2_VIDEO_START_CODE = 0x0100
	_SLICE_START_LOWER _MP2_VIDEO_START_CODE = 0x0101
	_SLICE_START_UPPER _MP2_VIDEO_START_CODE = 0x01af
	_USER_DATA_START   _MP2_VIDEO_START_CODE = 0x01b2
	_SEQUENCE_HEADER   _MP2_VIDEO_START_CODE = 0x01b3
	_SEQUENCE_ERROR    _MP2_VIDEO_START_CODE = 0x01b4
	_EXTENSION_START   _MP2_VIDEO_START_CODE = 0x01b5
	_SEQUENCE_END      _MP2_VIDEO_START_CODE = 0x01b7
	_GROUP_START       _MP2_VIDEO_START_CODE = 0x01b8
	_SYSTEM_START      _MP2_VIDEO_START_CODE = 0x01b9
)

type mpeg2Handler struct {
	bInit bool
}

func (h *mpeg2Handler) readSequenceHeader(r *common.BsReader) {
	r.ReadAndAssertBits(32, int(_SEQUENCE_HEADER), "sequence_header_code not match")
	hSize := r.ReadBits(12)
	vSize := r.ReadBits(12)
	aspectRatio := r.ReadBits(4)
	frame_rate_code := r.ReadBits(4)
	bit_rate_value := r.ReadBits(18)
	r.ReadBits(1)
	fmt.Println(fmt.Sprintf("Resolution: %d x %d, aspect ratio: %d, frame_rate_code: %d, bit_rate_value: %d", hSize, vSize, aspectRatio, frame_rate_code, bit_rate_value))
}

func (h *mpeg2Handler) Feed(unit common.CmUnit) {
	cmBuf, _ := unit.GetBuf().(common.CmBuf)
	buf := cmBuf.GetBuf()
	r := common.GetBufferReader(buf)
	fmt.Println(buf[:20])

	h.readSequenceHeader(&r)
	panic("End here")
}

func MPEG2VideoHandler() utils.DataHandler {
	return &mpeg2Handler{bInit: false}
}
