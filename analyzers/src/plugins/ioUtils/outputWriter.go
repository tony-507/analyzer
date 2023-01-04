package ioUtils

import (
	"encoding/json"

	"github.com/tony-507/analyzers/src/common"
	"github.com/tony-507/analyzers/src/logs"
)

type IWriter interface {
	setup(ioWriterParam)
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

func (w *OutputWriter) setCallback(callback common.RequestHandler) {
	w.callback = callback
}

func (w *OutputWriter) setParameter(m_parameter string) {
	var writerParam ioWriterParam
	if err := json.Unmarshal([]byte(m_parameter), &writerParam); err != nil {
		panic(err)
	}
	switch writerParam.OutputType {
	case _OUTPUT_FILE:
		w.impl = getFileWriter()
	}
	w.impl.setup(writerParam)
}

func (w *OutputWriter) setResource(loader *common.ResourceLoader) {}

func (w *OutputWriter) startSequence() {
	w.logger.Log(logs.INFO, "Output writer is started")
	w.isRunning = true

	common.Listen_msg(w.callback, w.name, 0x10)
}

func (w *OutputWriter) endSequence() {
	w.logger.Log(logs.INFO, "Output writer end sequence")
	w.isRunning = false
	w.impl.stop()
	w.logger.Log(logs.INFO, "Output writer impl stopped")
	eosUnit := common.MakeReqUnit(w.name, common.EOS_REQUEST)
	common.Post_request(w.callback, w.name, eosUnit)
}

func (w *OutputWriter) deliverUnit(unit common.CmUnit) {
	w.impl.processUnit(unit)
}

func (w *OutputWriter) fetchUnit() common.CmUnit {
	return nil
}

func (w *OutputWriter) deliverStatus(unit common.CmUnit) {
	w.impl.processControl(unit)
}

func GetOutputWriter(name string) common.Plugin {
	rv := OutputWriter{name: name, isRunning: false, logger: logs.CreateLogger(name)}
	return common.CreatePlugin(
		name,
		false,
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
