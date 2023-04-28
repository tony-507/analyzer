package ioUtils

import (
	"encoding/json"
	"os"

	"github.com/tony-507/analyzers/src/common"
)

const (
	TS_PKT_SIZE int = 188
)

type IReader interface {
	setup()                            // Set up reader
	startRecv() error                  // Start receiver
	stopRecv() error                   // Stop receiver
	dataAvailable(*common.IOUnit) bool // Get next unit of data
}

type inputReaderPlugin struct {
	logger      common.Log
	callback    common.RequestHandler
	impl        IReader
	isRunning   bool
	outputQueue []common.CmUnit
	outCnt      int
	name        string
	skipCnt     int
	maxInCnt    int
	rawDataFile *os.File
}

func (ir *inputReaderPlugin) StartSequence() {
	ir.isRunning = true

	err := ir.impl.startRecv()
	if err != nil {
		panic(err)
	}
}

func (ir *inputReaderPlugin) EndSequence() {
	ir.logger.Info("Ending sequence, fetch count = %d", ir.outCnt)
	ir.isRunning = false
	err := ir.impl.stopRecv()
	if err != nil {
		panic(err)
	}
	eosUnit := common.MakeReqUnit(ir.name, common.EOS_REQUEST)
	common.Post_request(ir.callback, ir.name, eosUnit)
}

func (ir *inputReaderPlugin) SetCallback(callback common.RequestHandler) {
	ir.callback = callback
}

func (ir *inputReaderPlugin) SetParameter(m_parameter string) {
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
		fname := "output/rawBuffer"
		f, err := os.Create(fname)
		ir.rawDataFile = f
		if err != nil {
			ir.logger.Error("Fail to create and open %s: %s", fname, err.Error())
		}
	}

	ir.outCnt = 0
	srcType := "unknown"

	switch param.Source {
	case _SOURCE_DUMMY:
		srcType = "dummy"
		ir.impl = &dummyReader{}
	case _SOURCE_FILE:
		srcType = "file"
		ir.impl = FileReader(ir.name, param.FileInput.Fname)
	case _SOURCE_UDP:
		srcType = "UDP"
		ir.impl = udpReader(&param.UdpInput, ir.name)
	}

	ir.logger.Info("%s reader created", srcType)

	ir.impl.setup()
}

func (ir *inputReaderPlugin) SetResource(loader *common.ResourceLoader) {}

func (ir *inputReaderPlugin) DeliverUnit(unit common.CmUnit) {
	if ir.isRunning {
		ir.start()
		reqUnit := common.MakeReqUnit(ir.name, common.DELIVER_REQUEST)
		common.Post_request(ir.callback, ir.name, reqUnit)
	}
}

func (ir *inputReaderPlugin) DeliverStatus(unit common.CmUnit) {}

func (ir *inputReaderPlugin) start() {
	// Here, we will keep delivering until EOS is signaled
	newUnit := common.IOUnit{}
	if ir.maxInCnt != 0 && ir.impl.dataAvailable(&newUnit) {
		if newUnit.Buf != nil {
			ir.outCnt += 1
			ir.maxInCnt -= 1
			ir.outputQueue = append(ir.outputQueue, &newUnit)
			reqUnit := common.MakeReqUnit(ir.name, common.FETCH_REQUEST)
			common.Post_request(ir.callback, ir.name, reqUnit)
		}
	} else {
		ir.EndSequence()
	}
}

func (ir *inputReaderPlugin) FetchUnit() common.CmUnit {
	var rv common.CmUnit

	if len(ir.outputQueue) != 0 && ir.skipCnt <= 0 {
		rv = ir.outputQueue[0]
	}

	if len(ir.outputQueue) == 1 {
		ir.outputQueue = make([]common.CmUnit, 0)
	} else {
		ir.outputQueue = ir.outputQueue[1:]
	}

	ir.skipCnt -= 1

	if ir.rawDataFile != nil && rv != nil {
		buf, _ := rv.GetBuf().([]byte)
		ir.rawDataFile.Write(buf)
	}

	return rv
}

func (ir *inputReaderPlugin) IsRoot() bool {
	return true
}

func (ir *inputReaderPlugin) Name() string {
	return ir.name
}

func InputReader(name string) common.IPlugin {
	rv := inputReaderPlugin{name: name, logger: common.CreateLogger(name)}
	return &rv
}
