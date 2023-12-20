package tsdemux

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/tony-507/analyzers/src/common"
	"github.com/tony-507/analyzers/src/common/logging"
	"github.com/tony-507/analyzers/src/tttKernel"
)

type IDemuxPipe interface {
	getDuration() int
	getOutputUnit() common.CmUnit
	processUnit([]byte, int) error
	start()
	stop()
}

type tsDemuxerPlugin struct {
	logger      logging.Log
	callback    tttKernel.RequestHandler
	impl        IDemuxPipe       // Actual demuxing operation
	control     *demuxController // Controller to handle demuxer internal state
	isRunning   int              // Counting channels, similar to waitGroup
	name        string
}

func (m_pMux *tsDemuxerPlugin) SetCallback(callback tttKernel.RequestHandler) {
	m_pMux.callback = callback
}

func (m_pMux *tsDemuxerPlugin) SetParameter(m_parameter string) {
	var demuxParam demuxParams
	if err := json.Unmarshal([]byte(m_parameter), &demuxParam); err != nil {
		panic(err)
	}
	// Do this here to prevent seg fault
	m_pMux.control = getControl()
	pipeType := "unknown"
	switch demuxParam.Mode {
	case _DEMUX_DUMMY:
		pipeType = "Dummy"
		impl := getDummyPipe(m_pMux)
		m_pMux.impl = &impl
	case _DEMUX_FULL:
		pipeType = "Demux"
		impl := getDemuxPipe(m_pMux, m_pMux.control, m_pMux.name)
		m_pMux.impl = &impl
	}
	m_pMux.logger.Info("%s pipe is started", pipeType)
	m_pMux._setup()
}

func (m_pMux *tsDemuxerPlugin) SetResource(resourceLoader *tttKernel.ResourceLoader) {
	m_pMux.control.setResource(resourceLoader)
}

func (m_pMux *tsDemuxerPlugin) _setup() {
	m_pMux.logger = logging.CreateLogger(m_pMux.name)
	m_pMux.isRunning = 0
}

func (m_pMux *tsDemuxerPlugin) StartSequence() {
	m_pMux.impl.start()
}

func (m_pMux *tsDemuxerPlugin) EndSequence() {
	m_pMux.logger.Info("Shutting down handlers")
	m_pMux.control.stop()

	m_pMux.impl.stop()

	eosUnit := common.MakeReqUnit(m_pMux.name, common.EOS_REQUEST)
	tttKernel.Post_request(m_pMux.callback, m_pMux.name, eosUnit)
}

func (m_pMux *tsDemuxerPlugin) FetchUnit() common.CmUnit {
	rv := m_pMux.impl.getOutputUnit()
	m_pMux.control.outputUnitFetched()
	return rv
}

func (m_pMux *tsDemuxerPlugin) DeliverUnit(inUnit common.CmUnit, intputId string) {
	m_pMux.control.inputReceived()

	// Perform demuxing on the received TS packet
	buf := common.GetBytesInBuf(inUnit)
	procErr := m_pMux.impl.processUnit(buf, m_pMux.control.getInputCount())
	if procErr != nil {
		m_pMux.logger.Error("At pkt#%d, %s", m_pMux.control.getInputCount(), procErr)
	}
}

func (m_pMux *tsDemuxerPlugin) DeliverStatus(unit common.CmUnit) {}

func (m_pMux *tsDemuxerPlugin) PrintInfo(sb *strings.Builder) {
	sb.WriteString(fmt.Sprintf("\tDuration: %f", float64(m_pMux.impl.getDuration()) / 27000000))
	m_pMux.control.printInfo(sb)
}

func (m_pMux *tsDemuxerPlugin) Name() string {
	return m_pMux.name
}

func (m_pMux *tsDemuxerPlugin) outputReady() {
	reqUnit := common.MakeReqUnit(m_pMux.name, common.FETCH_REQUEST)
	tttKernel.Post_request(m_pMux.callback, m_pMux.name, reqUnit)
}

func (m_pMux *tsDemuxerPlugin) getOutDir() string {
	return m_pMux.control.resourceLoader.Query("outDir", nil)
}

func TsDemuxer(name string) tttKernel.IPlugin {
	rv := tsDemuxerPlugin{
		name: name,
	}
	return &rv
}
