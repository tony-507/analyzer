package video

import (
	"fmt"

	"github.com/tony-507/analyzers/src/common/io"
	"github.com/tony-507/analyzers/src/common/logging"
	"github.com/tony-507/analyzers/src/plugins/dataHandler/utils"
	"github.com/tony-507/analyzers/src/plugins/dataHandler/video/h264"
	"github.com/tony-507/analyzers/src/tttKernel"
)

const (
	_H264_START_CODE_PREFIX = 0x000001
)

type h264Handler struct{
	logger logging.Log
	inCnt int
	sqp    h264.SequenceParameterSet
}

// 7.3.1
func (h *h264Handler) readNalUnit(r *io.BsReader, data *utils.VideoDataStruct) {
	r.ReadBits(1) // forbidden_zero_bit
	r.ReadBits(2) // nal_ref_idc
	nal_unit_type := r.ReadBits(5)
	nalUnitHeaderBytes := 1
	if nal_unit_type == 14 || nal_unit_type == 20 {
		nalUnitHeaderBytes += 3 
	}
	// HACK: Speed up processing by assuming coded slice always present as the last NAL unit
	if nal_unit_type <= 5 || nal_unit_type == 19 {
		h264.ReadSlice(r, data)
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
	r := io.GetBufferReader(rbsp)
	seiMsgs := []h264.Sei{}
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
		sei := h264.Sei{
			PayloadType: payloadType,
			PayloadSize: payloadSize,
			Buffer: r.GetRemainedBuffer()[:payloadSize],
		}
		seiMsgs = append(seiMsgs, sei)
		r.ReadBits(payloadSize * 8)
		if r.PeekBits(1) == 1 {
			r.ReadBits(1)
			for len(r.GetRemainedBuffer()) > 0 && r.PeekBits(1) == 0 {
				r.ReadBits(1)
			}
		}
	}

	for _, sei := range seiMsgs {
		switch sei.PayloadType {
		case 1:
			reader := io.GetBufferReader(sei.Buffer)
			picTiming := h264.ParsePicTiming(&reader, h.sqp)
			for _, clock := range picTiming.Clocks {
				data.TimeCode = clock.Tc
			}
		}
	}
}

func (h *h264Handler) isLeadingOrTrailingZeros(r *io.BsReader) bool {
	return (len(r.GetRemainedBuffer()) > 4 && r.PeekBits(32) != _H264_START_CODE_PREFIX) &&
	(len(r.GetRemainedBuffer()) > 3 && r.PeekBits(24) != _H264_START_CODE_PREFIX)
}

func (h *h264Handler) Feed(unit tttKernel.CmUnit, newData *utils.ParsedData) error {
	buf := tttKernel.GetBytesInBuf(unit)
	r := io.GetBufferReader(buf)
	nalCnt := 0
	data := newData.GetVideoData()

	h.inCnt++

	// Annex B.1.1
	for {
		for h.isLeadingOrTrailingZeros(&r) {
			if r.ReadBits(8) != 0 {
				h.logger.Error("leading_zero_8bits is not zero at pkt #%d", h.inCnt)
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
		logger: logging.CreateLogger(fmt.Sprintf("H264_%d", pid)),
		inCnt: 0,
		sqp: h264.CreateSequenceParameterSet(),
	}
}