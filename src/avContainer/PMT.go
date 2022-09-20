package avContainer

import (
	"fmt"

	"github.com/tony-507/analyzers/src/common"
)

type Descriptor struct {
	Tag     int
	Content string
}

type DataStream struct {
	StreamPid  int
	StreamType PKT_TYPE
	StreamDesc []Descriptor
}

type PMT struct {
	PktCnt     int
	PmtPid     int
	tableId    int
	ProgNum    int
	Version    int
	curNextIdr bool
	ProgDesc   []Descriptor
	Streams    []DataStream
	crc32      int
}

func PMTReadyForParse(r *common.BsReader) bool {
	// Peek the length of the table and compare with current buffer size
	(*r).SetMarker()
	pFieldLen := (*r).ReadBits(8)
	pktLen := pFieldLen + 1
	// First check here to see if we can safely get the section length
	if pktLen > 180 {
		return false
	}
	(*r).ReadBits(pFieldLen*8 + 8)
	pktLen += (*r).ReadBits(10)
	(*r).GoToMarker()
	return pktLen <= (*r).GetSize()
}

func ParsePMT(buf []byte, PmtPid int, cnt int) PMT {
	r := common.GetBufferReader(buf)

	pFieldLen := r.ReadBits(8)
	r.ReadBits(pFieldLen * 8)

	tableId := r.ReadBits(8)
	if r.ReadBits(1) != 1 {
		panic("Section syntax indicator of PMT is not set to 1")
	}
	if r.ReadBits(1) != 0 {
		panic("Private bits of PMT is not set to 0")
	}
	if r.ReadBits(2) != 3 {
		panic("Reserved bits of PMT is not set to all 1s")
	}
	if r.ReadBits(2) != 0 {
		panic("Unused bits of PMT is not set to all 0s")
	}
	sectionLen := r.ReadBits(10)

	ProgNum := r.ReadBits(16)
	if r.ReadBits(2) != 3 {
		panic("Reserved bits of PMT is not set to all 1s")
	}
	Version := r.ReadBits(5)
	curNextIdr := r.ReadBits(1)
	r.ReadBits(16) // section number and last section number

	sectionLen -= 9
	if r.ReadBits(3) != 7 {
		panic("Reserved bits of PMT is not set to all 1s")
	}
	// pcr_pid := r.ReadBits(13)
	r.ReadBits(13)
	if r.ReadBits(4) != 15 {
		panic("Reserved bits of PMT is not set to all 1s")
	}
	r.ReadBits(2) // Program info unused bits
	progInfoLen := r.ReadBits(10)
	sectionLen -= 4 + progInfoLen

	ProgDesc := make([]Descriptor, 0)
	for {
		if progInfoLen <= 0 {
			break
		}
		desc := _readDescriptor(&r, &progInfoLen)
		ProgDesc = append(ProgDesc, desc)
	}

	Streams := make([]DataStream, 0)
	for {
		if sectionLen == 0 {
			break
		}
		StreamType := r.ReadBits(8)
		if r.ReadBits(3) != 7 {
			panic("Reserved bits of PMT is not set to all 1s")
		}
		StreamPid := r.ReadBits(13)
		if r.ReadBits(4) != 15 {
			panic("Reserved bits of PMT is not set to all 1s")
		}
		if r.ReadBits(2) != 0 {
			panic("Unused bits of PMT is not set to all 0s")
		}
		esInfoLen := r.ReadBits(10)
		sectionLen -= 5 + esInfoLen

		StreamDesc := make([]Descriptor, 0)
		for {
			if esInfoLen <= 0 {
				break
			}
			desc := _readDescriptor(&r, &esInfoLen)
			StreamDesc = append(StreamDesc, desc)
		}
		Streams = append(Streams, DataStream{StreamPid, _resolveStreamType(StreamType), StreamDesc})
	}

	return PMT{PktCnt: cnt, PmtPid: PmtPid, tableId: tableId, ProgNum: ProgNum, Version: Version, curNextIdr: curNextIdr != 0, ProgDesc: ProgDesc, Streams: Streams, crc32: -1}
}

func (t *PMT) Pretty() {
	fmt.Println("\nAt pkt#", t.PktCnt)
	fmt.Println("Program", t.ProgNum)
	for idx, stream := range t.Streams {
		fmt.Println(" Stream", idx, ": type", _verboseType(stream.StreamType), " with pid", stream.StreamPid)
	}
}

func _readDescriptor(r *common.BsReader, l *int) Descriptor {
	Tag := (*r).ReadBits(8)
	descLen := (*r).ReadBits(8)
	Content := (*r).ReadHex(descLen)
	*l -= descLen + 2
	return Descriptor{Tag, Content}
}

func _resolveStreamType(typeId int) PKT_TYPE {
	switch typeId {
	case 2:
		return PKT_MPEG2
	case 4:
		return PKT_MPEG1
	default:
		return PKT_UNKNOWN
	}
}

func _verboseType(pktType PKT_TYPE) string {
	switch pktType {
	case PKT_MPEG2:
		return "MPEG 2 video"
	case PKT_MPEG1:
		return "MPEG 1 audio"
	default:
		return "Unknown"
	}
}
