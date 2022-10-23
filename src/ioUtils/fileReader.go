package ioUtils

import (
	"os"
	"strings"
	"path"
	"io"

	"github.com/tony-507/analyzers/src/common"
)

type INPUT_TYPE int

const (
	INPUT_UNKNOWN INPUT_TYPE = 0x00
	INPUT_TS      INPUT_TYPE = 0x01
	INPUT_MXF     INPUT_TYPE = 0x02
	INPUT_PCAP    INPUT_TYPE = 0x03
	INPUT_M2V     INPUT_TYPE = 0x10
)

type fileReader struct {
	fname   string
	fHandle *os.File
	ext     INPUT_TYPE
}

func (fr *fileReader) setup() {
	ext := strings.ToLower(path.Ext(fr.fname)[1:])
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

func (fr *fileReader) startRecv() {
	// Parse the data to have it in the form of a TS packet
	if fr.ext != INPUT_TS {
		panic("Input file type not supported. Please check the extension")
	}

	// Open the file and start reading
	fHandle, err := os.Open(fr.fname)
	check(err)
	fr.fHandle = fHandle
}

func (fr *fileReader) stopRecv() {
	fr.fHandle.Close()
}

func (fr *fileReader) dataAvailable(unit *common.IOUnit) bool {
	buf := make([]byte, TS_PKT_SIZE)
	n, err := fr.fHandle.Read(buf)
	if err == io.EOF {
		return false
	} else {
		check(err)
	}
	if n < TS_PKT_SIZE {
		return false
	}
	unit.IoType = 3
	unit.Buf = buf
	return true
}
