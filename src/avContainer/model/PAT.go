package model

// Parsing and writing of PAT
// Known issue:
// * Not support PAT with size > 1 TS packet

import (
	"errors"

	"github.com/tony-507/analyzers/src/common"
)

type PAT struct {
	PktCnt     int
	tableId    int
	tableIdExt int
	Version    int
	curNextIdr bool
	ProgramMap map[int]int
	Crc32      int
}

func PATReadyForParse(buf []byte) bool {
	r := common.GetBufferReader(buf)

	pFieldLen := r.ReadBits(8)
	// First check here to see if we can safely get the section length
	if len(r.GetRemainedBuffer()) < 2 {
		return false
	}
	r.ReadBits(pFieldLen*8 + 8 + 6)

	pktLen := pFieldLen + 3
	pktLen += r.ReadBits(10)

	return pktLen <= r.GetSize()
}

func ParsePAT(buf []byte, cnt int) (PAT, error) {
	r := common.GetBufferReader(buf)
	pFieldLen := r.ReadBits(8)
	r.ReadBits(pFieldLen * 8)

	tableId := r.ReadBits(8)
	if r.ReadBits(1) != 1 {
		err := errors.New("Section syntax indicator of PAT is not set to 1")
		return PAT{}, err
	}
	if r.ReadBits(1) != 0 {
		err := errors.New("Private bits of PAT is not set to 0")
		return PAT{}, err
	}
	if r.ReadBits(2) != 3 {
		err := errors.New("Reserved bits of PAT is not set to all 1s")
		return PAT{}, err
	}
	if r.ReadBits(2) != 0 {
		err := errors.New("Unused bits of PAT is not set to all 0s")
		return PAT{}, err
	}
	sectionLen := r.ReadBits(10)

	tableIdExt := r.ReadBits(16)
	if r.ReadBits(2) != 3 {
		err := errors.New("Reserved bits of PAT is not set to all 1s")
		return PAT{}, err
	}
	Version := r.ReadBits(5)
	curNextIdr := r.ReadBits(1)
	r.ReadBits(16) // section number and last section number

	sectionLen -= 9
	ProgramMap := make(map[int]int)
	for {
		if sectionLen <= 0 {
			break
		}
		progNum := r.ReadBits(16)
		if r.ReadBits(3) != 7 {
			err := errors.New("Reserved bits of PAT is not set to all 1s")
			return PAT{}, err
		}
		pid := r.ReadBits(13)
		ProgramMap[pid] = progNum
		sectionLen -= 4
	}
	if sectionLen < 0 {
		// Protection
		panic("Something wrong with section length")
	}
	crc32 := r.ReadBits(4)
	return PAT{PktCnt: cnt, tableId: tableId, tableIdExt: tableIdExt, Version: Version, curNextIdr: curNextIdr != 0, ProgramMap: ProgramMap, Crc32: crc32}, nil
}

func CreatePAT(tableId int, tableIdExt int, version int, curNextIdr bool, programMap map[int]int, crc32 int) PAT {
	return PAT{PktCnt: 0, tableId: tableId, tableIdExt: tableIdExt, Version: version, curNextIdr: curNextIdr, ProgramMap: programMap, Crc32: crc32}
}
