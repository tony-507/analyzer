package fileReader

/*
 * A class for handling file containing raw binary data.
 *
 * The reader
 * - extracts UDP payload based on file type
 * - extracts application payload based on configured application-layer protocols
 */

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/tony-507/analyzers/src/common"
	"github.com/tony-507/analyzers/src/plugins/ioUtils/def"
	"github.com/tony-507/analyzers/src/plugins/ioUtils/protocol"
)

type FileReaderStruct struct {
	logger      common.Log
	fname       string
	fHandle     *os.File
	config      def.IReaderConfig
	bufferQueue []def.ParseResult
}

func (fr *FileReaderStruct) Setup(config def.IReaderConfig) {
	fr.config = config
}

func (fr *FileReaderStruct) StartRecv() error {
	// Open the file and start reading
	fHandle, err := os.Open(fr.fname)
	if err != nil {
		return err
	}
	fr.fHandle = fHandle

	var buf []byte
	splitRes := strings.Split(fr.fname, ".")
	ext := splitRes[len(splitRes) - 1]

	if ext == "pcap" {
		buf = []byte{}
		pcap, err := pcapFile(fr.fHandle, fr.logger)
		if err != nil {
			return err
		}
		for {
			rawBuf, err := pcap.getBuffer()
			if err != nil {
				return err
			}
			if len(rawBuf) == 0 {
				fr.logger.Info("No more buffer from pcap")
				break
			}
			buf = append(buf, rawBuf...)
		}
	} else {
		stat, err := fr.fHandle.Stat()
		if err != nil {
			return errors.New(fmt.Sprintf("Fail to retrieve file stat: %s", err.Error()))
		}
		buf = make([]byte, stat.Size())
		_, err = fr.fHandle.Read(buf)
		if err != nil {
			return errors.New(fmt.Sprintf("Fail to read buffer: %s", err.Error()))
		}
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

func FileReader(name string, fname string) def.IReader {
	rv := &FileReaderStruct{
		logger: common.CreateLogger(name),
		fname: fname,
		config: def.IReaderConfig{},
		bufferQueue: []def.ParseResult{},
	}
	return rv
}
