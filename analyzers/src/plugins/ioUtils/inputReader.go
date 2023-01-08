package ioUtils

import (
	"encoding/json"

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
	callback    common.RequestHandler
	impl        IReader
	isRunning   bool
	outputQueue []common.CmUnit
	outCnt      int
	name        string
	skipCnt     int
	maxInCnt    int
}

func (ir *InputReader) startSequence() {
	ir.logger.Log(logs.INFO, "File reader is started")
	ir.isRunning = true

	ir.impl.startRecv()
}

func (ir *InputReader) endSequence() {
	ir.logger.Log(logs.INFO, "Stopping file reader, fetch count = %d", ir.outCnt)
	ir.isRunning = false
	ir.impl.stopRecv()
	eosUnit := common.MakeReqUnit(ir.name, common.EOS_REQUEST)
	common.Post_request(ir.callback, ir.name, eosUnit)
}

func (ir *InputReader) setCallback(callback common.RequestHandler) {
	ir.callback = callback
}

func (ir *InputReader) setParameter(m_parameter string) {
	var param ioReaderParam
	if err := json.Unmarshal([]byte(m_parameter), &param); err != nil {
		panic(err)
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

	if param.DumpRawInput {
		buf := common.MakeSimpleBuf([]byte{})
		buf.SetField("pid", -1, false)
		buf.SetField("addPid", true, false)
		buf.SetField("type", 3, false)
		unit := common.MakeStatusUnit(0x10, buf)
		common.Post_status(ir.callback, ir.name, unit)
	}

	ir.outCnt = 0

	switch param.Source {
	case _SOURCE_DUMMY:
		ir.impl = &dummyReader{}
	case _SOURCE_FILE:
		ir.impl = &fileReader{fname: param.FileInput.Fname}
	}

	ir.impl.setup()
}

func (ir *InputReader) setResource(loader *common.ResourceLoader) {}

func (ir *InputReader) deliverUnit(unit common.CmUnit) {
	for ir.isRunning {
		ir.start()
	}
}

func (ir *InputReader) deliverStatus(unit common.CmUnit) {}

func (ir *InputReader) start() {
	// Here, we will keep delivering until EOS is signaled
	newUnit := common.IOUnit{}
	if ir.maxInCnt != 0 && ir.impl.dataAvailable(&newUnit) {
		ir.outCnt += 1
		ir.maxInCnt -= 1
		ir.outputQueue = append(ir.outputQueue, &newUnit)
		reqUnit := common.MakeReqUnit(ir.name, common.FETCH_REQUEST)
		common.Post_request(ir.callback, ir.name, reqUnit)
	} else {
		ir.endSequence()
	}
}

func (ir *InputReader) fetchUnit() common.CmUnit {
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

	rv := common.MakeIOUnit(nil, 0, -1)
	return rv
}

func GetInputReader(name string) common.Plugin {
	rv := InputReader{name: name, logger: logs.CreateLogger(name)}
	return common.CreatePlugin(
		name,
		true,
		rv.setCallback,
		rv.setParameter,
		rv.setResource,
		rv.startSequence,
		rv.deliverUnit,
		rv.deliverStatus,
		rv.fetchUnit,
		rv.endSequence,
	)
}
