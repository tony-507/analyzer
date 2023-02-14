package ioUtils

import (
	"encoding/json"

	"github.com/tony-507/analyzers/src/common"
)

type IWriter interface {
	setup(ioWriterParam)
	stop()
	processUnit(common.CmUnit)
	processControl(common.CmUnit)
}

type OutputWriter struct {
	logger    common.Log
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
	outType := "unknown"
	switch writerParam.OutputType {
	case _OUTPUT_FILE:
		outType = "file"
		w.impl = getFileWriter(w.name)
	}
	w.logger.Info("%s writer is started", outType)
	w.impl.setup(writerParam)
}

func (w *OutputWriter) setResource(loader *common.ResourceLoader) {}

func (w *OutputWriter) startSequence() {
	w.isRunning = true

	common.Listen_msg(w.callback, w.name, 0x10)
}

func (w *OutputWriter) endSequence() {
	w.logger.Info("Ending sequence")
	w.isRunning = false
	w.impl.stop()
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
	rv := OutputWriter{name: name, isRunning: false, logger: common.CreateLogger(name)}
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
