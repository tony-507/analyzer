package ioUtils

import (
	"common"
	"os"
	"path"
	"strings"
)

const (
	TS_PKT_SIZE int = 188
)

type INPUT_TYPE int

const (
	INPUT_UNKNOWN INPUT_TYPE = 0x00
	INPUT_TS      INPUT_TYPE = 0x01
	INPUT_MXF     INPUT_TYPE = 0x02
	INPUT_PCAP    INPUT_TYPE = 0x03
	INPUT_M2V     INPUT_TYPE = 0x10
)

func check(err error) {
	if err != nil {
		panic(err)
	}
}

type FileReader struct {
	fHandle     []byte
	outputQueue []common.CmUnit
	ext         INPUT_TYPE
	outCnt      int
}

func (fr *FileReader) _setup(fname string) {
	fr.outCnt = 0

	// Open the file and start reading
	inBuf, err := os.ReadFile(fname)
	fr.fHandle = inBuf
	check(err)

	ext := strings.ToLower(path.Ext(fname)[1:])
	switch ext {
	case "ts":
		fr.ext = INPUT_TS
	case "tp":
		fr.ext = INPUT_TS
	case "mxf":
		fr.ext = INPUT_MXF
	case "pcap":
		fr.ext = INPUT_PCAP
	default:
		fr.ext = INPUT_UNKNOWN
	}
}

func (fr *FileReader) DeliverUnit(unit common.IOUnit) {
	// Parse the data to have it in the form of a TS packet
	if fr.ext != INPUT_TS {
		panic("Input file type not supported. Please check the extension")
	}
	processedUnit := common.IOUnit{IoType: 1, Buf: fr.fHandle[(fr.outCnt * TS_PKT_SIZE):((fr.outCnt + 1) * TS_PKT_SIZE)]}
	fr.outputQueue = append(fr.outputQueue, processedUnit)

	fr.outCnt += 1
}

func (fr *FileReader) FetchUnit() common.CmUnit {
	if len(fr.outputQueue) != 0 {
		rv := fr.outputQueue[0]
		if len(fr.outputQueue) == 1 {
			fr.outputQueue = make([]common.CmUnit, 0)
		} else {
			fr.outputQueue = fr.outputQueue[1:]
		}
		return rv
	}

	rv := common.IOUnit{IoType: 0, Buf: nil}
	return rv
}

func (fr *FileReader) DataAvailable() bool {
	// Check if there is still data to read
	fSize := len(fr.fHandle)
	return fr.outCnt*TS_PKT_SIZE < fSize
}

// Wrapper to skip initialization line outside package
func GetReader(fname string) FileReader {
	rv := FileReader{}
	rv._setup(fname)
	return rv
}
