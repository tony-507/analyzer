package fileReader

/*
 * A class for handling file containing raw binary data.
 *
 * The reader
 * - extracts UDP payload based on file type
 * - extracts application payload based on configured application-layer protocols
 */

import (
	"os"
	"strings"
	"sync"

	"github.com/tony-507/analyzers/src/common/logging"
	"github.com/tony-507/analyzers/src/common/protocol"
	"github.com/tony-507/analyzers/src/plugins/ioUtils/def"
)

type fileHandler interface {
	getBuffer() ([]byte, error)
	tick() (int64, bool)
}

type FileReaderStruct struct {
	logger      logging.Log
	fname       string
	fHandle     *os.File
	config      def.IReaderConfig
	bufferQueue []protocol.ParseResult
	running     bool
	mtx         sync.Mutex
	wg          sync.WaitGroup
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

	fr.running = true
	fr.wg.Add(1)
	go fr.worker()

	return nil
}

func (fr *FileReaderStruct) worker() {
	splitRes := strings.Split(fr.fname, ".")
	ext := splitRes[len(splitRes) - 1]
	var handler fileHandler

	switch ext {
	case "pcap":
		handler = pcapFile(fr.fHandle, fr.logger)
	default:
		handler = binaryDataFile(fr.fHandle)
	}

	for fr.running {
		buf, err := handler.getBuffer()
		if err != nil {
			panic(err)
		}
		if len(buf) == 0 {
			fr.logger.Info("No more buffer from file")
			break
		}

		results := protocol.ParseWithParsers(fr.config.Parsers, &protocol.ParseResult{Buffer: buf})
		if realtime, ok := handler.tick(); ok {
			for i := range results {
				if results[i].Fields == nil {
					results[i].Fields = map[string]int64{}
				}
				results[i].Fields["realtimeInUs"] = realtime
			}
		}

		fr.mtx.Lock()
		fr.bufferQueue = append(fr.bufferQueue, results...)
		fr.mtx.Unlock()
	}

	fr.running = false
	fr.wg.Done()
}

func (fr *FileReaderStruct) StopRecv() error {
	fr.running = false
	fr.wg.Wait()
	return fr.fHandle.Close()
}

func (fr *FileReaderStruct) DataAvailable() (protocol.ParseResult, bool) {
	fr.mtx.Lock()
	defer fr.mtx.Unlock()
	if len(fr.bufferQueue) > 0 {
		buf := fr.bufferQueue[0]
		if len(fr.bufferQueue) > 1 {
			fr.bufferQueue = fr.bufferQueue[1:]
		} else {
			fr.bufferQueue = []protocol.ParseResult{}
		}
		return buf, true
	} else if fr.running {
		return protocol.EmptyResult(), true
	}
	return protocol.ParseResult{}, false
}

func FileReader(name string, fname string) def.IReader {
	rv := &FileReaderStruct{
		logger: logging.CreateLogger(name),
		fname: fname,
		config: def.IReaderConfig{},
		bufferQueue: []protocol.ParseResult{},
	}
	return rv
}
