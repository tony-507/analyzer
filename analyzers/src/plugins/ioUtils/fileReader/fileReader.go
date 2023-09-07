package fileReader

import (
	"errors"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/tony-507/analyzers/src/common"
	"github.com/tony-507/analyzers/src/plugins/ioUtils/def"
	"github.com/tony-507/analyzers/src/plugins/ioUtils/protocol"
)

type INPUT_TYPE int

const (
	INPUT_UNKNOWN INPUT_TYPE = 0x00
	INPUT_TS      INPUT_TYPE = 0x01
	INPUT_MXF     INPUT_TYPE = 0x02
	INPUT_PCAP    INPUT_TYPE = 0x03
	INPUT_M2V     INPUT_TYPE = 0x10
)

type FileReaderStruct struct {
	logger      common.Log
	fname       string
	fHandle     *os.File
	ext         INPUT_TYPE
	config      def.IReaderConfig
	bufferQueue []def.ParseResult
}

func (fr *FileReaderStruct) Setup(config def.IReaderConfig) {
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
	fr.config = config
}

func (fr *FileReaderStruct) StartRecv() error {
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

	stat, err := fr.fHandle.Stat()
	if err != nil {
		return errors.New(fmt.Sprintf("Fail to retrieve file stat: %s", err.Error()))
	}
	buf := make([]byte, stat.Size())
	_, err = fr.fHandle.Read(buf)
	if err != nil {
		return errors.New(fmt.Sprintf("Fail to read buffer: %s", err.Error()))
	}
	fr.bufferQueue = append(fr.bufferQueue, protocol.ParseWithParsers(fr.config.Protocols, buf)...)

	return nil
}

func (fr *FileReaderStruct) StopRecv() error {
	return fr.fHandle.Close()
}

func (fr *FileReaderStruct) DataAvailable(unit *common.IOUnit) bool {
	if len(fr.bufferQueue) > 0 {
		unit.IoType = 3
		unit.Id = -1
		unit.Buf = fr.bufferQueue[0].GetBuffer()
		if len(fr.bufferQueue) > 1 {
			fr.bufferQueue = fr.bufferQueue[1:]
		} else {
			fr.bufferQueue = []def.ParseResult{}
		}
		return true
	}
	return false
}

func (fr *FileReaderStruct) GetType() INPUT_TYPE {
	return fr.ext
}

func FileReader(name string, fname string) def.IReader {
	rv := &FileReaderStruct{
		logger: common.CreateLogger(name),
		fname: fname,
		config: def.IReaderConfig{},
		bufferQueue: []def.ParseResult{},
	}
	return rv
}
