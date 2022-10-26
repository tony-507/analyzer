package ioUtils

import (
	"github.com/tony-507/analyzers/src/common"
	"github.com/tony-507/analyzers/src/logs"
)

const (
	TS_PKT_SIZE int = 188
)

type IReader interface {
	setup()                            // Set up reader
	startRecv()                        // Start receiver
	stopRecv()                         // Stop receiver
	dataAvailable(*common.IOUnit) bool // Get next unit of data
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}

type InputReader struct {
	logger      logs.Log
	impl        IReader
	outputQueue []common.CmUnit
	outCnt      int
	name        string
	skipCnt     int
	maxInCnt    int
}

func (ir *InputReader) StartSequence() {
	ir.logger.Log(logs.INFO, "File reader is started")
	ir.impl.startRecv()
}

func (ir *InputReader) EndSequence() {
	ir.logger.Log(logs.INFO, "Stopping file reader, fetch count = ", ir.outCnt)
	ir.impl.stopRecv()
}

func (ir *InputReader) SetParameter(m_parameter interface{}) {
	param, isInputParam := m_parameter.(IOReaderParam)
	if !isInputParam {
		panic("File reader param has unknown format")
	}
	if param.SkipCnt > 0 {
		ir.skipCnt = param.SkipCnt
	} else {
		ir.skipCnt = 0
	}
	if param.MaxInCnt > 0 {
		ir.maxInCnt = param.MaxInCnt
	} else {
		ir.maxInCnt = -1
	}

	ir.outCnt = 0

	switch param.Source {
	case SOURCE_DUMMY:
		ir.impl = &dummyReader{}
	case SOURCE_FILE:
		ir.impl = &fileReader{fname: param.FileInput.Fname}
	}

	ir.impl.setup()
}

func (ir *InputReader) DeliverUnit(unit common.CmUnit) common.CmUnit {
	newUnit := common.IOUnit{}
	if ir.maxInCnt != 0 && ir.impl.dataAvailable(&newUnit) {
		ir.outCnt += 1
		ir.maxInCnt -= 1
		ir.outputQueue = append(ir.outputQueue, newUnit)
		reqUnit := common.MakeReqUnit(ir.name, common.FETCH_REQUEST)
		return reqUnit
	} else {
		ir.EndSequence()
		reqUnit := common.MakeReqUnit(ir.name, common.EOS_REQUEST)
		return reqUnit
	}
}

func (ir *InputReader) FetchUnit() common.CmUnit {
	if len(ir.outputQueue) != 0 && ir.skipCnt <= 0 {
		rv := ir.outputQueue[0]
		if len(ir.outputQueue) == 1 {
			ir.outputQueue = make([]common.CmUnit, 0)
		} else {
			ir.outputQueue = ir.outputQueue[1:]
		}
		return rv
	}

	ir.skipCnt -= 1

	rv := common.IOUnit{IoType: 0, Buf: nil}
	return rv
}

// Wrapper to skip initialization line outside package
func GetReader(name string) InputReader {
	return InputReader{name: name, logger: logs.CreateLogger("inputReader")}
}
