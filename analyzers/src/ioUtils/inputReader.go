package ioUtils

import (
	"io"
	"os"
	"path"
	"strings"

	"github.com/tony-507/analyzers/src/common"
	"github.com/tony-507/analyzers/src/logs"
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
	fHandle     *os.File
	logger	logs.Log
	outputQueue []common.CmUnit
	ext         INPUT_TYPE
	outCnt      int
	name        string
}

func (fr *FileReader) StartSequence() {
	fr.logger.Log(logs.INFO, "File reader is started")
}

func (fr *FileReader) EndSequence() {
	fr.fHandle.Close()
}

func (fr *FileReader) SetParameter(m_parameter interface{}) {
	param, isInputParam := m_parameter.(IOReaderParam)
	if !isInputParam {
		panic("File reader param has unknown format")
	}
	fr._setup(param.Fname)
}

func (fr *FileReader) _setup(fname string) {
	fr.outCnt = 0

	// Open the file and start reading
	fHandle, err := os.Open(fname)
	check(err)
	fr.fHandle = fHandle

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

func (fr *FileReader) DeliverUnit(unit common.CmUnit) common.CmUnit {
	// Parse the data to have it in the form of a TS packet
	if fr.ext != INPUT_TS {
		panic("Input file type not supported. Please check the extension")
	}

	if fr.DataAvailable() {
		fr.outCnt += 1
		reqUnit := common.MakeReqUnit(fr.name, common.FETCH_REQUEST)
		return reqUnit
	} else {
		reqUnit := common.MakeReqUnit(fr.name, common.EOS_REQUEST)
		return reqUnit
	}
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
	processedUnit := common.IOUnit{IoType: 1, Buf: buf}
	fr.outputQueue = append(fr.outputQueue, processedUnit)
	return true
}

// Wrapper to skip initialization line outside package
func GetReader(name string) FileReader {
	return FileReader{name: name, logger: logs.CreateLogger("inputReader")}
}