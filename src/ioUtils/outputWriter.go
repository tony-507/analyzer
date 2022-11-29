package ioUtils

import (
	"github.com/tony-507/analyzers/src/common"
	"github.com/tony-507/analyzers/src/logs"
	"github.com/tony-507/analyzers/src/resources"
)

type IWriter interface {
	setup(IOWriterParam)
	stop()
	processUnit(common.CmUnit)
}

type OutputWriter struct {
	logger    logs.Log
	callback  common.PostRequestHandler
	impl      IWriter
	name      string
	isRunning bool
}

func (w *OutputWriter) SetCallback(callback common.PostRequestHandler) {
	w.callback = callback
}

func (w *OutputWriter) SetParameter(m_parameter interface{}) {
	writerParam, isWriterParam := m_parameter.(IOWriterParam)
	if !isWriterParam {
		panic("Writer param has unknown format")
	}
	switch writerParam.OutputType {
	case OUTPUT_DUMMY:
		w.impl = &DummyWriter{}
	case OUTPUT_FILE:
		w.impl = GetFileWriter()
	}
	w.impl.setup(writerParam)
}

func (w *OutputWriter) SetResource(loader *resources.ResourceLoader) {}

func (w *OutputWriter) StartSequence() {
	w.logger.Log(logs.INFO, "Output writer is started")
	w.isRunning = true
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

func GetOutputWriter(name string) OutputWriter {
	return OutputWriter{name: name, isRunning: false, logger: logs.CreateLogger("OutputWriter")}
}
