package ioUtils

import (
	"encoding/json"

	"github.com/tony-507/analyzers/src/common"
)

type IWriter interface {
	setup(ioWriterParam) error
	stop()
	processUnit(common.CmUnit)
	processControl(common.CmUnit)
}

type outputWriterPlugin struct {
	logger    common.Log
	callback  common.RequestHandler
	impl      IWriter
	name      string
	isRunning bool
}

func (w *outputWriterPlugin) SetCallback(callback common.RequestHandler) {
	w.callback = callback
}

func (w *outputWriterPlugin) SetParameter(m_parameter string) {
	var writerParam ioWriterParam
	if err := json.Unmarshal([]byte(m_parameter), &writerParam); err != nil {
		panic(err)
	}
	outType := "unknown"
	// TODO: Add an implementation here
	w.logger.Info("%s writer is started", outType)
	err := w.impl.setup(writerParam)
	if err != nil {
		panic(err)
	}

	common.Listen_msg(w.callback, w.name, 0x10)
}

func (w *outputWriterPlugin) SetResource(loader *common.ResourceLoader) {}

func (w *outputWriterPlugin) StartSequence() {
	w.isRunning = true
}

func (w *outputWriterPlugin) EndSequence() {
	w.logger.Info("Ending sequence")
	w.isRunning = false
	w.impl.stop()
	eosUnit := common.MakeReqUnit(w.name, common.EOS_REQUEST)
	common.Post_request(w.callback, w.name, eosUnit)
}

func (w *outputWriterPlugin) DeliverUnit(unit common.CmUnit) {
	w.impl.processUnit(unit)
}

func (w *outputWriterPlugin) FetchUnit() common.CmUnit {
	return nil
}

func (w *outputWriterPlugin) DeliverStatus(unit common.CmUnit) {
	w.impl.processControl(unit)
}

func (w *outputWriterPlugin) IsRoot() bool {
	return false
}

func (w *outputWriterPlugin) Name() string {
	return w.name
}

func OutputWriter(name string) common.IPlugin {
	rv := outputWriterPlugin{name: name, isRunning: false, logger: common.CreateLogger(name)}
	return &rv
}
