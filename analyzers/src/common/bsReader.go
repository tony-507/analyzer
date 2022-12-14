package common

import (
	"encoding/hex"
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

func (br *BsReader) ReadHex(n int) string {
	rv := ""
	for i := 0; i < n; i++ {
		rv += " " + string(strconv.FormatInt(int64(br.ReadBits(8)), 16))
	}
	return rv[1:] // Remove leading space
}

func (br *BsReader) ReadChar(n int) string {
	rv := ""
	for i := 0; i < n; i++ {
		rv += string(strconv.FormatInt(int64(br.ReadBits(8)), 16))
	}
	bs, err := hex.DecodeString(rv)
	if err != nil {
		panic(err)
	}
	return string(bs)
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
