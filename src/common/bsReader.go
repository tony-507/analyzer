package common

import (
	"strconv"
)

type BsReader struct {
	rawBs        []byte
	pos          int
	offset       int
	markerPos    int
	markerOffset int
}

func GetBufferReader(rawBs []byte) BsReader {
	return BsReader{rawBs, 0, 8, -1, -1}
}

func (br *BsReader) GetCurByte() int {
	return int(br.rawBs[br.pos])
}

func (br *BsReader) GetRemainedBuffer() []byte {
	// Seems like Go returns this by reference, so returning the buf directly may result in seg fault
	return append(make([]byte, 0), br.rawBs[br.pos:]...)
}

func (br *BsReader) GetSize() int {
	return len(br.rawBs)
}

func (br *BsReader) GetPos() int {
	return br.pos
}

func (br *BsReader) GetOffset() int {
	return br.offset
}

func (br *BsReader) AddData(r BsReader) {
	br.rawBs = append(br.rawBs, r.GetRemainedBuffer()...)
}

func (br *BsReader) SetMarker() {
	// Set up a marker at a bit position that can be returned later
	br.markerPos = br.pos
	br.markerOffset = br.offset
}

func (br *BsReader) GoToMarker() {
	// Sanity check: If no marker set, panic
	if br.markerPos == -1 {
		panic("Marker is not set")
	}
	br.pos = br.markerPos
	br.offset = br.markerOffset
	br.markerPos = -1
	br.markerOffset = -1
}

func (br *BsReader) ReadHex(n int) string {
	rv := ""
	for i := 0; i < n; i++ {
		rv += " " + string(strconv.FormatInt(int64(br.ReadBits(8)), 16))
	}
	return rv
}

func (br *BsReader) ReadBits(n int) int {
	rv := 0
	if n >= br.offset {
		mask := getMask(br.offset)
		rv = int(br.rawBs[br.pos]) & mask
		br.pos += 1
		n -= br.offset
		br.offset = 8
		if n > 0 {
			// Read remaining bits from the next byte
			rv = (rv << n) + br.ReadBits(n)
		}
	} else if n > 0 {
		// Within same byte
		mask := getMask(br.offset) - getMask(br.offset-n)
		rv = (int(br.rawBs[br.pos]) & mask) >> (br.offset - n)
		br.offset -= n
		// Normalize offset back to [0,8]
		if br.offset < 0 {
			for {
				if br.offset > 0 {
					break
				}
				br.offset += 8
			}
		}
		if br.offset == 8 {
			br.pos += 1
		}
	}
	return rv
}

func getMask(n int) int {
	rv := -1
	switch n {
	case 0:
	case 8:
		rv = 0xff
		break
	default:
		rv = (1 << n) - 1
	}
	return rv
}
