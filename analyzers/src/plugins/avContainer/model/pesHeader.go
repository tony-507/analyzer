package model

import (
	"errors"
	"fmt"

	"github.com/tony-507/analyzers/src/common"
)

type OptionalHeader struct {
	scrambled   bool // If true, the remaining would not be parsed
	dataAligned bool // If false, the remaining would not be parsed
	length      int  // Length of the piece
	pts         int  // -1 means not present
	dts         int  // same as dts
}

type PESHeader struct {
	streamId       int
	sectionLen     int
	optionalHeader OptionalHeader
}

func (h *PESHeader) GetSectionLength() int {
	return h.sectionLen
}

func (h *PESHeader) GetPts() int {
	return h.optionalHeader.pts
}

func (h *PESHeader) GetDts() int {
	return h.optionalHeader.dts
}

// Parse optional header and return its length
func ParseOptionalHeader(buf []byte) (OptionalHeader, error) {
	// Initialize optional values here
	pts := -1
	dts := -1
	r := common.GetBufferReader(buf)

	// Begin reading
	if r.ReadBits(2) != 2 {
		err := errors.New("optional PES header marker bits not match")
		return OptionalHeader{}, err
	}
	scrambled := r.ReadBits(2) != 0
	r.ReadBits(1) // Priority
	dataAligned := r.ReadBits(1) != 0
	r.ReadBits(1) // Copyright
	r.ReadBits(1) // Original/ copy
	pts_dts_flag := r.ReadBits(2)
	r.ReadBits(1) // ESCR flag
	r.ReadBits(1) // ES rate flag
	r.ReadBits(1) // DSM trick mode flag
	r.ReadBits(1) // Additional copy info flag
	r.ReadBits(1) // CRC flag
	r.ReadBits(1) // Extension flag
	headerLen := r.ReadBits(8)

	// Optional fields
	remained := headerLen

	// PTS DTS handling
	switch pts_dts_flag {
	case 3:
		sync := r.ReadBits(4)
		if sync != 3 {
			errMsg := fmt.Sprintf("PTS first four bits not match: sync=%d, flag=%d", sync, pts_dts_flag)
			err := errors.New(errMsg)
			return OptionalHeader{}, err
		}
		pts = r.ReadBits(3)
		r.ReadBits(1)
		pts = (pts << 15) + r.ReadBits(15)
		r.ReadBits(1)
		pts = (pts << 15) + r.ReadBits(15)
		r.ReadBits(1)

		sync = r.ReadBits(4)
		if sync != 1 {
			errMsg := fmt.Sprintf("DTS first four bits not match: sync=%d, flag=%d", sync, pts_dts_flag)
			err := errors.New(errMsg)
			return OptionalHeader{}, err
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
			errMsg := fmt.Sprintf("PTS first four bits not match: sync=%d, flag=%d", sync, pts_dts_flag)
			err := errors.New(errMsg)
			return OptionalHeader{}, err
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
		errMsg := fmt.Sprintf("Forbidden timestamp flag: flag=%d", pts_dts_flag)
		err := errors.New(errMsg)
		return OptionalHeader{}, err
	}

	r.ReadBits(remained * 8)
	remained = 0

	return OptionalHeader{scrambled: scrambled, dataAligned: dataAligned, length: headerLen + 3, pts: pts, dts: dts}, nil
}

func ParsePESHeader(buf []byte) (PESHeader, int, error) {
	logger := common.CreateLogger("PesParser")

	r := common.GetBufferReader(buf)
	headerLen := 0

	if r.ReadBits(24) != 0x000001 {
		err := errors.New("PES prefix start code not match")
		logger.Error("%v", r)
		return PESHeader{}, 0, err
	}
	streamId := r.ReadBits(8)
	pesLen := r.ReadBits(16)

	headerLen += 6

	// TODO: May not have optional header
	optionalHeader, err := ParseOptionalHeader(r.GetRemainedBuffer())
	if err != nil {
		logger.Error("%v", r)
		errMsg := fmt.Sprintf("%s\nReader status: (%d, %d)", err.Error(), r.GetPos(), r.GetOffset())
		err = errors.New(errMsg)
		return PESHeader{}, 0, err
	}
	headerLen += optionalHeader.length

	if pesLen != 0 {
		pesLen -= optionalHeader.length
	} else {
		pesLen = len(r.GetRemainedBuffer()) - optionalHeader.length
	}

	return PESHeader{streamId, pesLen, optionalHeader}, headerLen, nil
}

func CreatePESHeader(streamId int, pesLen int, optionalHeader OptionalHeader) PESHeader {
	return PESHeader{streamId: streamId, sectionLen: pesLen, optionalHeader: optionalHeader}
}

func CreateOptionalPESHeader(length int, pts int, dts int) OptionalHeader {
	return OptionalHeader{scrambled: false, dataAligned: true, length: length, pts: pts, dts: dts}
}
