package tsdemux

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	"github.com/tony-507/analyzers/src/common"
	"github.com/tony-507/analyzers/src/logs"
)

type IDemuxPipe interface {
	processUnit([]byte, int)
	getDuration() int
	getProgramNumber(int) int
	readyForFetch() bool
	getOutputUnit() common.CmUnit
}

type PKT_TYPE int

func GetVersion(buf []byte) int {
	r := common.GetBufferReader(buf)
	pFieldLen := r.ReadBits(8)
	// Protection on failure to get version
	if pFieldLen+7 > r.GetSize()-4 {
		return -1
	}
	r.ReadBits(pFieldLen*8 + 8 + 6 + 10 + 16 + 2)
	return r.ReadBits(5)
}

const (
	PKT_UNKNOWN PKT_TYPE = -1
	PKT_PAT     PKT_TYPE = 0
	PKT_SDT     PKT_TYPE = 1
	PKT_PMT     PKT_TYPE = 2
	PKT_SCTE    PKT_TYPE = 3
	PKT_VIDEO   PKT_TYPE = 17
	PKT_MPEG2   PKT_TYPE = 18
	PKT_AVC     PKT_TYPE = 19
	PKT_HEVC    PKT_TYPE = 20
	PKT_AUDIO   PKT_TYPE = 33
	PKT_MPEG1   PKT_TYPE = 34
	PKT_AC3     PKT_TYPE = 35
	PKT_AAC     PKT_TYPE = 36
	PKT_NULL    PKT_TYPE = 8191
)

type TsDemuxer struct {
	logger    logs.Log
	callback  common.RequestHandler
	impl      IDemuxPipe       // Actual demuxing operation
	control   *demuxController // Controller to handle demuxer internal state
	isRunning int              // Counting channels, similar to waitGroup
	pktCnt    int              // The index of currently fed packet
	name      string
	wg        sync.WaitGroup
}

func (m_pMux *TsDemuxer) setCallback(callback common.RequestHandler) {
	m_pMux.callback = callback
}

func (m_pMux *TsDemuxer) setParameter(m_parameter string) {
	var demuxParam demuxParams
	if err := json.Unmarshal([]byte(m_parameter), &demuxParam); err != nil {
		panic(err)
	}
	// Do this here to prevent seg fault
	m_pMux.control = getControl()
	switch demuxParam.Mode {
	case _DEMUX_DUMMY:
		impl := getDummyPipe(m_pMux)
		m_pMux.impl = &impl
	case _DEMUX_FULL:
		impl := getDemuxPipe(m_pMux.control)
		m_pMux.impl = &impl
	}
	m_pMux._setup()
}

func (m_pMux *TsDemuxer) setResource(resourceLoader *common.ResourceLoader) {
	m_pMux.control.setResource(resourceLoader)
}

func (m_pMux *TsDemuxer) _setup() {
	m_pMux.logger = logs.CreateLogger("TsDemuxer")
	m_pMux.pktCnt = 1
	m_pMux.isRunning = 0

	m_pMux.wg.Add(1)
	go m_pMux._setupMonitor()
}

// Demuxer monitor, run as a Goroutine to monitor demuxer's status
// Currently only check if demuxer gets stuck
func (m_pMux *TsDemuxer) _setupMonitor() {
	defer m_pMux.wg.Done()
	m_pMux.control.monitor()
}

func (m_pMux *TsDemuxer) startSequence() {
	m_pMux.logger.Log(logs.INFO, "TSDemuxer has started")
}

func (m_pMux *TsDemuxer) endSequence() {
	// Fetch all remaining units
	for m_pMux.impl.readyForFetch() {
		m_pMux.control.outputUnitAdded()
		reqUnit := common.MakeReqUnit(m_pMux.name, common.FETCH_REQUEST)
		common.Post_request(m_pMux.callback, m_pMux.name, reqUnit)
	}
	m_pMux.logger.Log(logs.INFO, "Shutting down handlers")
	m_pMux.control.stop()
	m_pMux.control.printSummary(m_pMux.impl.getDuration())
	m_pMux.wg.Wait()
	eosUnit := common.MakeReqUnit(m_pMux.name, common.EOS_REQUEST)
	common.Post_request(m_pMux.callback, m_pMux.name, eosUnit)
}

func (m_pMux *TsDemuxer) fetchUnit() common.CmUnit {
	rv := m_pMux.impl.getOutputUnit()
	errMsg := ""

	if cmBuf, isCmBuf := rv.GetBuf().(common.CmBuf); isCmBuf {
		id, _ := rv.GetField("id").(int)
		// Skip the following for PSI
		if id == 1 {
			if field, hasField := cmBuf.GetField("progNum"); hasField {
				progNum, _ := field.(int)
				// Stamp PCR here
				clk := m_pMux.control.updateSrcClk(progNum)

				if field, hasField = cmBuf.GetField("pktCnt"); hasField {
					curCnt, _ := field.(int)
					id, isInt := rv.GetField("id").(int)
					if !isInt {
						errMsg = fmt.Sprintf("Invalid id in data unit: %v", rv)
					} else {
						pcr, _ := clk.requestPcr(id, curCnt)
						cmBuf.SetField("PCR", pcr, false)
						if field, hasField = cmBuf.GetField("DTS"); hasField {
							dts, _ := field.(int)
							cmBuf.SetField("Delay", dts-pcr/300, false)
						}
					}
				} else {
					errMsg = "Unable to get pktCnt"
				}
			} else {
				errMsg = "Unable to get progNum"
			}
		}
	}

	if errMsg != "" {
		common.Throw_error(m_pMux.callback, m_pMux.name, errors.New(fmt.Sprintf("[TSDemuxer::FetchUnit] %s.", errMsg)))
	}

	m_pMux.control.outputUnitFetched()

	return rv
}

func (m_pMux *TsDemuxer) deliverUnit(inUnit common.CmUnit) {
	m_pMux.control.inputReceived()

	// Perform demuxing on the received TS packet
	inBuf, _ := inUnit.GetBuf().([]byte)
	m_pMux.impl.processUnit(inBuf, m_pMux.pktCnt)

	m_pMux.pktCnt += 1

	// Clean up status
	for _, status := range m_pMux.control.StatusList {
		common.Post_status(m_pMux.callback, m_pMux.name, status)
	}
	m_pMux.control.StatusList = make([]common.CmUnit, 0)

	// Start fetching after clock is ready
	if m_pMux.impl.readyForFetch() {
		m_pMux.control.outputUnitAdded()
		reqUnit := common.MakeReqUnit(m_pMux.name, common.FETCH_REQUEST)
		common.Post_request(m_pMux.callback, m_pMux.name, reqUnit)
	}
}

func (m_pMux *TsDemuxer) deliverStatus(unit common.CmUnit) {}

func GetTsDemuxer(name string) common.Plugin {
	rv := TsDemuxer{name: name}
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
