package video

import (
	"fmt"

	"github.com/tony-507/analyzers/src/common"
	"github.com/tony-507/analyzers/src/plugins/dataHandler/utils"
	"github.com/tony-507/analyzers/src/plugins/dataHandler/video/h264"
)

const (
	_H264_START_CODE_PREFIX = 0x000001
)

type h264Handler struct{
	logger common.Log
	inCnt int
	sqp    h264.SequenceParameterSet
}

// 7.3.1
func (h *h264Handler) readNalUnit(r *common.BsReader, data *utils.VideoDataStruct) {
	r.ReadBits(1) // forbidden_zero_bit
	r.ReadBits(2) // nal_ref_idc
	nal_unit_type := r.ReadBits(5)
	nalUnitHeaderBytes := 1
	if nal_unit_type == 14 || nal_unit_type == 20 {
		nalUnitHeaderBytes += 3 
	}
	rbsp := []byte{}
	for {
		// Annex B.2
		if len(r.GetRemainedBuffer()) > 2 && (r.PeekBits(24) == 0 || r.PeekBits(24) == 1) {
			break
		}
		if len(r.GetRemainedBuffer()) > 2 && r.PeekBits(24) == 0x000003 {
			rbsp = append(rbsp, r.GetRemainedBuffer()[:2]...)
			r.ReadBits(16)
			r.ReadBits(8) // emulation_prevention_three_bytes
		} else if len(r.GetRemainedBuffer()) > 0 {
			rbsp = append(rbsp, byte(r.ReadBits(8)))
		} else {
			break
		}
	}
	switch nal_unit_type {
	case 6:
		h.readSEI(rbsp, data)
	case 7:
		h.sqp = h264.ParseSequenceParameterSet(rbsp)
	default:
		// Unhandled
	}
}

func (h *h264Handler) readSEI(rbsp []byte, data *utils.VideoDataStruct) {
	r := common.GetBufferReader(rbsp)
	// Last byte is rbsp_trailing_bits
	for len(r.GetRemainedBuffer()) > 1 {
		payloadType := 0
		payloadSize := 0
		
		for r.PeekBits(8) == 0xff {
			r.ReadBits(8)
			payloadType += 255
		}
		payloadType += r.ReadBits(8)
		
		for r.PeekBits(8) == 0xff {
			r.ReadBits(8)
			payloadSize += 255
		}
		payloadSize += r.ReadBits(8)

		switch payloadType {
		case 1:
			picTiming := h264.ParsePicTiming(&r, h.sqp)
			for _, clock := range picTiming.Clocks {
				data.TimeCode = clock.Tc
			}
		default:
			r.ReadBits(payloadSize * 8)
		}

		if r.GetOffset() != 8 {
			r.ReadBits(r.GetOffset())
		}
	}
}

func (h *h264Handler) isLeadingOrTrailingZeros(r *common.BsReader) bool {
	return (len(r.GetRemainedBuffer()) > 4 && r.PeekBits(32) != _H264_START_CODE_PREFIX) &&
	(len(r.GetRemainedBuffer()) > 3 && r.PeekBits(24) != _H264_START_CODE_PREFIX)
}

func (h *h264Handler) Feed(unit common.CmUnit, newData *utils.ParsedData) error {
	cmBuf, _ := unit.GetBuf().(common.CmBuf)
	buf := cmBuf.GetBuf()
	r := common.GetBufferReader(buf)
	nalCnt := 0
	data := newData.GetVideoData()

	h.inCnt++

	// Annex B.1.1
	for {
		for h.isLeadingOrTrailingZeros(&r) {
			if r.ReadBits(8) != 0 {
				h.logger.Error("leading_zero_8bits is not zero")
			}
		}
		if len(r.GetRemainedBuffer()) == 0 {
			break
		}
		if r.PeekBits(24) != _H264_START_CODE_PREFIX {
			r.ReadBits(8)
		}
		r.ReadBits(24)
		nalCnt++
		h.readNalUnit(&r, data)
		for h.isLeadingOrTrailingZeros(&r) {
			if r.ReadBits(8) != 0 {
				h.logger.Error("trailing_zero_8bits is not zero")
			}
		}
	}
	return nil
}

func H264VideoHandler(pid int) utils.DataHandler {
	return &h264Handler{
		logger: common.CreateLogger(fmt.Sprintf("H264_%d", pid)),
		inCnt: 0,
		sqp: h264.CreateSequenceParameterSet(),
	}
}