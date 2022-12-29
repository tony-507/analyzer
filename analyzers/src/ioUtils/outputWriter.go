package ioUtils

import (
	"encoding/json"

	"github.com/tony-507/analyzers/src/common"
	"github.com/tony-507/analyzers/src/logs"
	"github.com/tony-507/analyzers/src/resources"
)

type IWriter interface {
	setup(IOWriterParam)
	stop()
	processUnit(common.CmUnit)
	processControl(common.CmUnit)
}

type OutputWriter struct {
	logger    logs.Log
	callback  common.RequestHandler
	impl      IWriter
	name      string
	isRunning bool
}

func (w *OutputWriter) SetCallback(callback common.RequestHandler) {
	w.callback = callback
}

func (w *OutputWriter) SetParameter(m_parameter string) {
	var writerParam IOWriterParam
	if err := json.Unmarshal([]byte(m_parameter), &writerParam); err != nil {
		panic(err)
	}
	switch writerParam.OutputType {
	case OUTPUT_FILE:
		w.impl = GetFileWriter()
	}
	w.impl.setup(writerParam)
}

func (w *OutputWriter) SetResource(loader *resources.ResourceLoader) {}

func (w *OutputWriter) StartSequence() {
	w.logger.Log(logs.INFO, "Output writer is started")
	w.isRunning = true

	common.Listen_msg(w.callback, w.name, 0x10)
}

func (w *OutputWriter) EndSequence() {
	w.logger.Log(logs.INFO, "Output writer end sequence")
	w.isRunning = false
	w.impl.stop()
	w.logger.Log(logs.INFO, "Output writer impl stopped")
	eosUnit := common.MakeReqUnit(w.name, common.EOS_REQUEST)
	common.Post_request(w.callback, w.name, eosUnit)
}

func (w *OutputWriter) DeliverUnit(unit common.CmUnit) {
	w.impl.processUnit(unit)
}

func (w *OutputWriter) FetchUnit() common.CmUnit {
	return nil
}

func (w *OutputWriter) DeliverStatus(unit common.CmUnit) {
	w.impl.processControl(unit)
}

func GetOutputWriter(name string) OutputWriter {
	return OutputWriter{name: name, isRunning: false, logger: logs.CreateLogger("OutputWriter")}
}
