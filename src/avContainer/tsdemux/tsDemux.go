package tsdemux

import (
	"sync"

	"github.com/tony-507/analyzers/src/common"
	"github.com/tony-507/analyzers/src/logs"
	"github.com/tony-507/analyzers/src/resources"
)

type IDemuxPipe interface {
	processUnit([]byte, int)
	getDuration() int
	getProgramNumber(int) int
	clockReady() bool
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
	logger         logs.Log
	impl           IDemuxPipe             // Actual demuxing operation
	control        *demuxController       // Controller to handle demuxer internal state
	progClkMap     map[int]*programSrcClk // progNum -> srcClk
	outputQueue    []common.IOUnit        // Outputs to other plugins
	isRunning      int                    // Counting channels, similar to waitGroup
	pktCnt         int                    // The index of currently fed packet
	resourceLoader *resources.ResourceLoader
	name           string
	wg             sync.WaitGroup
}

func (m_pMux *TsDemuxer) SetParameter(m_parameter interface{}) {
	demuxParam, isParam := m_parameter.(DemuxParams)
	if !isParam {
		panic("Unknown type received at TsDemuxer::SetParameter")
	}
	// Do this here to prevent seg fault
	m_pMux.control = getControl()
	switch demuxParam.Mode {
	case DEMUX_DUMMY:
		impl := getDummyPipe(m_pMux)
		m_pMux.impl = &impl
	case DEMUX_FULL:
		impl := getDemuxPipe(m_pMux)
		m_pMux.impl = &impl
	}
	m_pMux._setup()
}

func (m_pMux *TsDemuxer) SetResource(resourceLoader *resources.ResourceLoader) {
	m_pMux.resourceLoader = resourceLoader
}

func (m_pMux *TsDemuxer) _setup() {
	m_pMux.logger = logs.CreateLogger("TsDemuxer")
	m_pMux.progClkMap = make(map[int]*programSrcClk, 0)
	m_pMux.outputQueue = make([]common.IOUnit, 0)
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

func (m_pMux *TsDemuxer) _updateSrcClk(progNum int) *programSrcClk {
	_, hasKey := m_pMux.progClkMap[progNum]
	if !hasKey {
		clk := getProgramSrcClk(m_pMux)
		m_pMux.progClkMap[progNum] = &clk
	}
	return m_pMux.progClkMap[progNum]
}

func (m_pMux *TsDemuxer) StartSequence() {
	m_pMux.logger.Log(logs.INFO, "TSDemuxer has started")
}

func (m_pMux *TsDemuxer) EndSequence() {
	m_pMux.logger.Log(logs.INFO, "Shutting down handlers")
	m_pMux.control.stop()
	m_pMux.control.printSummary(m_pMux.impl.getDuration())
	m_pMux.wg.Wait()
}

func (m_pMux *TsDemuxer) FetchUnit() common.CmUnit {
	outLen := len(m_pMux.outputQueue)

	if outLen != 0 {
		rv := m_pMux.outputQueue[0]

		pesBuf, isPes := rv.Buf.(common.PesBuf)
		if isPes {
			progNum, _ := pesBuf.GetField("progNum").(int)

			// Stamp PCR here
			clk := m_pMux._updateSrcClk(progNum)

			curCnt, _ := pesBuf.GetField("pktCnt").(int)
			pcr, _ := clk.requestPcr(rv.Id, curCnt)
			pesBuf.SetPcr(pcr)

			rv.Buf = pesBuf
		} else {
			psiBuf, isPsiBuf := rv.Buf.(common.PsiBuf)
			if isPsiBuf {
				// Use clock of first program as we want relative time only
				clk := m_pMux._updateSrcClk(m_pMux.impl.getProgramNumber(0))

				curCnt, _ := psiBuf.GetField("pktCnt").(int)
				pcr, _ := clk.requestPcr(rv.Id, curCnt)
				psiBuf.SetPcr(pcr)

				rv.Buf = psiBuf
			}
		}

		if outLen > 1 {
			m_pMux.outputQueue = m_pMux.outputQueue[1:]
		} else {
			m_pMux.outputQueue = make([]common.IOUnit, 0)
		}

		m_pMux.control.outputUnitFetched()

		return rv
	}

	return nil
}

func (m_pMux *TsDemuxer) DeliverUnit(inUnit common.CmUnit) common.CmUnit {
	m_pMux.control.inputReceived()

	// Perform demuxing on the received TS packet
	inBuf, _ := inUnit.GetBuf().([]byte)
	m_pMux.impl.processUnit(inBuf, m_pMux.pktCnt)

	m_pMux.pktCnt += 1

	// Start fetching after clock is ready
	if m_pMux.impl.clockReady() {
		m_pMux.control.outputUnitAdded()
		reqUnit := common.MakeReqUnit(m_pMux.name, common.FETCH_REQUEST)
		return reqUnit
	} else {
		m_pMux.logger.Log(logs.INFO, "Demuxer returns null unit")
		return nil
	}
}

func GetTsDemuxer(name string) TsDemuxer {
	return TsDemuxer{name: name}
}
