package video

import (
	"fmt"

	"github.com/tony-507/analyzers/src/common"
	"github.com/tony-507/analyzers/src/common/io"
	"github.com/tony-507/analyzers/src/common/logging"
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
	logger logging.Log
	pid    int
	pesCnt int
	bInit  bool
}

func (h *mpeg2Handler) readSequenceHeader(r *io.BsReader) {
	hSize := r.ReadBits(12)
	vSize := r.ReadBits(12)
	aspectRatio := r.ReadBits(4)
	frame_rate_code := r.ReadBits(4)
	bit_rate_value := r.ReadBits(18)
	r.ReadBits(1)
	vbv_buffer_size_value := r.ReadBits(10)
	constrained_parameters_flag := r.ReadBits(1)
	if !h.bInit {
		h.logger.Trace(fmt.Sprintf("Resolution: %d x %d, aspect ratio: %d, frame_rate_code: %d, bit_rate_value: %d, vbv_buffer_size_value: %d, constrained_parameters_flag: %d",
			hSize, vSize, aspectRatio, frame_rate_code, bit_rate_value, vbv_buffer_size_value, constrained_parameters_flag))
		h.bInit = true
	}
	// intra_quantiser_matrix
	if r.ReadBits(1) != 0 {
		r.ReadBits(8 * 64)
	}
	// non_intra_quantiser_matrix
	if r.ReadBits(1) != 0 {
		r.ReadBits(8 * 64)
	}
}

func (h *mpeg2Handler) Feed(unit common.CmUnit, newData *utils.ParsedData) error {
	h.pesCnt += 1
	buf := common.GetBytesInBuf(unit)
	r := io.GetBufferReader(buf)

	nextBits := r.ReadBits(32)
	if nextBits == int(_SEQUENCE_HEADER) {
		h.readSequenceHeader(&r)
	}
	return nil
}

func MPEG2VideoHandler(pid int) utils.DataHandler {
	return &mpeg2Handler{logger: logging.CreateLogger(fmt.Sprintf("MP2_%d", pid)), pid: pid, pesCnt: 0, bInit: false}
}
