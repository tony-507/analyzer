package tsdemux

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/tony-507/analyzers/src/common"
	"github.com/tony-507/analyzers/src/utils"
)

type IDemuxCallback interface {
	outputReady()
}

type IDemuxPipe interface {
	getDuration() int
	getOutputUnit() common.CmUnit
	processUnit([]byte, int) error
}

type tsDemuxerPlugin struct {
	logger      common.Log
	callback    common.RequestHandler
	fileWriters map[string]map[int]utils.FileWriter
	impl        IDemuxPipe       // Actual demuxing operation
	control     *demuxController // Controller to handle demuxer internal state
	isRunning   int              // Counting channels, similar to waitGroup
	name        string
}

func (m_pMux *tsDemuxerPlugin) SetCallback(callback common.RequestHandler) {
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

func (m_pMux *tsDemuxerPlugin) SetResource(resourceLoader *common.ResourceLoader) {
	m_pMux.control.setResource(resourceLoader)
}

func (m_pMux *tsDemuxerPlugin) _setup() {
	m_pMux.logger = common.CreateLogger(m_pMux.name)
	m_pMux.isRunning = 0
}

func (m_pMux *tsDemuxerPlugin) StartSequence() {
	for _, fileType := range []string{"csv", "pes"} {
		m_pMux.fileWriters[fileType] = map[int]utils.FileWriter{}
	}
}

func (m_pMux *tsDemuxerPlugin) EndSequence() {
	m_pMux.logger.Info("Shutting down handlers")
	m_pMux.control.stop()

	for fileType, writers := range m_pMux.fileWriters {
		for pid, writer := range writers {
			err := writer.Close()
			if err != nil {
				m_pMux.logger.Error("Fail to close %s writer for pid %d: %s", fileType, pid, err.Error())
			}
		}
	}

	eosUnit := common.MakeReqUnit(m_pMux.name, common.EOS_REQUEST)
	common.Post_request(m_pMux.callback, m_pMux.name, eosUnit)
}

func (m_pMux *tsDemuxerPlugin) FetchUnit() common.CmUnit {
	rv := m_pMux.impl.getOutputUnit()
	errMsg := ""

	cmBuf := rv.GetBuf()
	if progNum, ok := common.GetBufFieldAsInt(cmBuf, "progNum"); ok {
		// Stamp PCR here
		clk := m_pMux.control.updateSrcClk(progNum)

		if curCnt, ok := common.GetBufFieldAsInt(cmBuf, "pktCnt"); ok {
			pid, _ := rv.GetField("id").(int)
			pcr, _ := clk.requestPcr(pid, curCnt)
			cmBuf.SetField("pcr", pcr, false)
			if dts, ok := common.GetBufFieldAsInt(cmBuf, "dts"); ok {
				cmBuf.SetField("delay", dts-pcr/300, false)
			}
			// Write output
			for _, fileType := range []string{"csv", "pes"} {
				shouldWrite := true
				if _, ok := m_pMux.fileWriters[fileType][pid]; !ok {
					shouldWrite = false
					outDir := "output"
					fname := fmt.Sprintf("%d.%s", pid, fileType)
					var fWriter utils.FileWriter
					switch fileType {
					case "csv":
						fWriter = utils.CsvWriter(outDir, fname)
					case "pes":
						fWriter = utils.RawWriter(outDir, fname)
					}
					m_pMux.fileWriters[fileType][pid] = fWriter
					if err := fWriter.Open(); err != nil {
						m_pMux.logger.Warn("Fail to open handler for %s: %s", fname, err.Error())
					} else {
						shouldWrite = true
					}
				}
				if shouldWrite {
					m_pMux.fileWriters[fileType][pid].Write(cmBuf)
				}
			}
		} else {
			errMsg = "Unable to get pktCnt"
		}
	} else {
		errMsg = "Unable to get progNum"
	}

	if errMsg != "" {
		common.Throw_error(m_pMux.callback, m_pMux.name, errors.New(fmt.Sprintf("[TSDemuxerPlugin::FetchUnit] %s.", errMsg)))
	}

	m_pMux.control.outputUnitFetched()

	return rv
}

func (m_pMux *tsDemuxerPlugin) DeliverUnit(inUnit common.CmUnit) {
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
	common.Post_request(m_pMux.callback, m_pMux.name, reqUnit)
}

func TsDemuxer(name string) common.IPlugin {
	rv := tsDemuxerPlugin{
		fileWriters: map[string]map[int]utils.FileWriter{},
		name: name,
	}
	return &rv
}
