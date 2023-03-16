package ioUtils

import (
	"errors"
	"io"
	"os"
	"path"
	"strings"

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

type fileReaderStruct struct {
	logger  common.Log
	fname   string
	fHandle *os.File
	ext     INPUT_TYPE
}

func (fr *fileReaderStruct) setup() {
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

func (fr *fileReaderStruct) startRecv() error {
	// Parse the data to have it in the form of a TS packet
	if fr.ext != INPUT_TS {
		return errors.New("Input file type not supported. Please check the extension")
	}

	// Open the file and start reading
	fHandle, err := os.Open(fr.fname)
	if err != nil {
		return err
	}
	fr.fHandle = fHandle

	return nil
}

func (fr *fileReaderStruct) stopRecv() error {
	return fr.fHandle.Close()
}

func (fr *fileReaderStruct) dataAvailable(unit *common.IOUnit) bool {
	buf := make([]byte, TS_PKT_SIZE)
	n, err := fr.fHandle.Read(buf)
	if err == io.EOF {
		return false
	} else {
		if err != nil {
			fr.logger.Error("Fail to read buffer: %s", err.Error())
		}
	}
	if n < TS_PKT_SIZE {
		return false
	}
	unit.IoType = 3
	unit.Id = -1
	unit.Buf = buf
	return true
}

func FileReader(name string, fname string) IReader {
	rv := &fileReaderStruct{logger: common.CreateLogger(name), fname: fname}
	return rv
}
