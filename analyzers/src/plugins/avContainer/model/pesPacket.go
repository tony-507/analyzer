package model

import (
	"errors"
	"fmt"

	"github.com/tony-507/analyzers/src/common"
)

type pesPacketStruct struct {
	pid               int
	header            common.CmBuf
	payload           []byte
	hasOptionalHeader bool
	sectionLen        int
	callback          pesHandle
}

func (p *pesPacketStruct) setBuffer(inBuf []byte, pktCnt int) error {
	buf := inBuf[:6]
	r := common.GetBufferReader(buf)
	p.header = common.MakeSimpleBuf(buf)

	p.header.SetField("pktCnt", pktCnt, false)
	p.header.SetField("size", -1, false)

	if r.ReadBits(24) != 0x000001 {
		return errors.New("PES prefix start code not match")
	}

	streamId := r.ReadBits(8)
	p.header.SetField("streamId", streamId, true)
	sectionLen := r.ReadBits(16) // Can be zero
	readLen := 6
	optionalHeaderLength := 0

	if streamId != 0x10111100 && streamId != 0x10111110 && streamId != 0x10111111 &&
		streamId != 0x11110000 && streamId != 0x11110001 && streamId != 0x11110010 &&
		streamId != 0x11111000 && streamId != 0x11111111 {
		var err error
		p.hasOptionalHeader = true
		optionalHeaderLength, err = p.readOptionalHeader(inBuf[6:])
		if err != nil {
			return err
		}
	} else {
		return errors.New("Special stream type not implemented for PES packet")
	}

	p.payload = inBuf[(6 + optionalHeaderLength):]

	if sectionLen == 0 {
		p.sectionLen = -1
	} else {
		p.sectionLen = readLen + sectionLen - 6 - optionalHeaderLength
	}

	return nil
}

func (p *pesPacketStruct) readOptionalHeader(buf []byte) (int, error) {
	r := common.GetBufferReader(buf)

	// Begin reading
	if r.ReadBits(2) != 2 {
		return 0, errors.New("optional PES header marker bits not match")
	}
	if r.ReadBits(2) != 0 {
		return 0, errors.New("PES packet is scrambled")
	}
	p.header.SetField("priority", r.ReadBits(1), true)
	r.ReadBits(1) // Data alignment indicator
	p.header.SetField("copyright", r.ReadBits(1), true)
	p.header.SetField("original", r.ReadBits(1), true)

	pts := -1
	dts := -1
	escr := -1
	esRate := -1

	ptsDtsIdr := r.ReadBits(2)
	hasEscr := r.ReadBits(1) != 0
	hasEsRate := r.ReadBits(1) != 0
	isDsmTrickMode := r.ReadBits(1) != 0
	hasAdditionalCopyInfo := r.ReadBits(1) != 0
	hasCrc := r.ReadBits(1) != 0
	hasExtension := r.ReadBits(1) != 0
	headerLen := r.ReadBits(8)

	remained := headerLen
	switch ptsDtsIdr {
	case 3:
		sync := r.ReadBits(4)
		if sync != 3 {
			errMsg := fmt.Sprintf("PTS first four bits not match: sync=%d, flag=%d", sync, ptsDtsIdr)
			return 0, errors.New(errMsg)
		}
		pts = r.ReadBits(3)
		r.ReadBits(1)
		pts = (pts << 15) + r.ReadBits(15)
		r.ReadBits(1)
		pts = (pts << 15) + r.ReadBits(15)
		r.ReadBits(1)

		sync = r.ReadBits(4)
		if sync != 1 {
			errMsg := fmt.Sprintf("DTS first four bits not match: sync=%d, flag=%d", sync, ptsDtsIdr)
			return 0, errors.New(errMsg)
		}
		dts = r.ReadBits(3)
		r.ReadBits(1)
		dts = (dts << 15) + r.ReadBits(15)
		r.ReadBits(1)
		dts = (dts << 15) + r.ReadBits(15)
		r.ReadBits(1)
		remained -= 10
	case 2:
		sync := r.ReadBits(4)
		if sync != 2 {
			errMsg := fmt.Sprintf("PTS first four bits not match: sync=%d, flag=%d", sync, ptsDtsIdr)
			return 0, errors.New(errMsg)
		}
		pts = r.ReadBits(3)
		r.ReadBits(1)
		pts = (pts << 15) + r.ReadBits(15)
		r.ReadBits(1)
		pts = (pts << 15) + r.ReadBits(15)
		dts = pts
		r.ReadBits(1)
		remained -= 5
	case 1:
		errMsg := fmt.Sprintf("Forbidden timestamp flag: flag=%d", ptsDtsIdr)
		return 0, errors.New(errMsg)
	}
	p.header.SetField("pts", pts, false)
	p.header.SetField("dts", dts, false)

	if hasEscr {
		r.ReadBits(2)
		escr = r.ReadBits(3)
		r.ReadBits(1)
		escr = (escr << 15) + r.ReadBits(15)
		r.ReadBits(1)
		escr = (escr << 15) + r.ReadBits(15)
		r.ReadBits(1)
		escr = escr*300 + r.ReadBits(9)
		r.ReadBits(1)
		remained -= 6
	}
	p.header.SetField("escr", escr, true)

	if hasEsRate {
		r.ReadBits(1)
		esRate = r.ReadBits(22) * 50
		r.ReadBits(1)
		remained -= 3
	}
	p.header.SetField("esRate", esRate, true)

	if isDsmTrickMode {
		control := r.ReadBits(3)
		switch control {
		case 0b000:
			// Fast forward
			r.ReadBits(2) // field_id
			r.ReadBits(1) // intra_slice_refresh
			r.ReadBits(2) // frequency_truncation
		case 0b001:
			// Slow motion
			r.ReadBits(5) // rep_cntrl
		case 0b010:
			// Freeze frame
			r.ReadBits(2) // field_id
			r.ReadBits(3)
		case 0b011:
			// Fast reverse
			r.ReadBits(2) // field_id
			r.ReadBits(1) // intra_slice_refresh
			r.ReadBits(2) // frequency_truncation
		case 0b100:
			// Slow reverse
			r.ReadBits(5) // rep_cntrl
		default:
			// Reserved
			r.ReadBits(5)
		}
		remained -= 1
	}

	if hasAdditionalCopyInfo {
		r.ReadBits(1)
		r.ReadBits(7) // additional_copy_info
		remained -= 1
	}

	if hasCrc {
		r.ReadBits(16) // previous_PES_packet_CRC
		remained -= 2
	}

	if hasExtension {
		// TODO
	}
	r.ReadBits(remained * 8)

	return headerLen + 3, nil
}

func (p *pesPacketStruct) Process() error {
	p.header.SetField("size", len(p.payload), false)
	p.callback.PesPacketReady(p.header, p.pid)
	return nil
}

func (p *pesPacketStruct) Append(buf []byte) {
	p.payload = append(p.payload, buf...)
}

func (p *pesPacketStruct) GetField(str string) (int, error) {
	return resolveHeaderField(p, str)
}

func (p *pesPacketStruct) GetName() string {
	return "PES packet"
}

func (p *pesPacketStruct) GetHeader() common.CmBuf {
	return p.header
}

func (p *pesPacketStruct) GetPayload() []byte {
	return p.payload
}

func (p *pesPacketStruct) Ready() bool {
	if p.sectionLen == -1 {
		return false
	} else {
		return len(p.payload) >= p.sectionLen
	}
}

func (p *pesPacketStruct) Serialize() []byte {
	// TODO
	return []byte{}
}

func PesPacket(callback pesHandle, buf []byte, pid int, pktCnt int, progNum int, streamType int) (DataStruct, error) {
	rv := &pesPacketStruct{pid: pid, hasOptionalHeader: false, payload: make([]byte, 0), callback: callback}
	err := rv.setBuffer(buf, pktCnt)

	rv.header.SetField("progNum", progNum, true)
	rv.header.SetField("streamType", streamType, true)

	return rv, err
}
