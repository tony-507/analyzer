package model

import (
	"github.com/tony-507/analyzers/src/common"
)

type Descriptor struct {
	Tag     int
	Content string
}

type DataStream struct {
	StreamPid  int
	StreamType int
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

func PMTReadyForParse(buf []byte) bool {
	// Peek the length of the table and compare with current buffer size
	r := common.GetBufferReader(buf)
	pFieldLen := r.ReadBits(8)
	// First check here to see if we can safely get the section length
	if len(r.GetRemainedBuffer()) < 2 {
		return false
	}
	r.ReadBits(pFieldLen*8 + 8 + 6)

	pmtLen := pFieldLen + 3
	pmtLen += r.ReadBits(10)

	return pmtLen <= r.GetSize()
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
		Streams = append(Streams, DataStream{StreamPid, StreamType, StreamDesc})
	}

	return PMT{PktCnt: cnt, PmtPid: PmtPid, tableId: tableId, ProgNum: ProgNum, Version: Version, curNextIdr: curNextIdr != 0, ProgDesc: ProgDesc, Streams: Streams, crc32: -1}
}

func _readDescriptor(r *common.BsReader, l *int) Descriptor {
	Tag := (*r).ReadBits(8)
	descLen := (*r).ReadBits(8)
	Content := (*r).ReadHex(descLen)
	*l -= descLen + 2
	return Descriptor{Tag, Content}
}

func CreatePMT(PmtPid int, tableId int, progNum int, version int, curNextIdr bool, progDesc []Descriptor, streams []DataStream, crc32 int) PMT {
	return PMT{PktCnt: 0, PmtPid: PmtPid, tableId: tableId, ProgNum: progNum, Version: version, curNextIdr: curNextIdr, ProgDesc: progDesc, Streams: streams, crc32: -1}
}
