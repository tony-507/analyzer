package tsdemux

import (
	"errors"
	"fmt"
	"strings"

	"github.com/tony-507/analyzers/src/common"
	"github.com/tony-507/analyzers/src/plugins/avContainer/model"
)

type tsDemuxPipe struct {
	logger          common.Log
	callback        IDemuxCallback
	control         *demuxController // Controller from demuxer
	dataStructs     map[int]model.DataStruct
	programRecords  map[int]int // PAT
	streamRecords   map[int]int // Stream pid => stream type
	streamTree      map[int]int // Stream pid => program number
	patVersion      int
	pmtVersions     map[int]int     // Program number => version
	outputQueue     []common.CmUnit // Outputs to other plugins
	scte35SplicePTS map[int][]int   // Program number => Splice PTS
}

func (m_pMux *tsDemuxPipe) _setup() {
	m_pMux.programRecords = make(map[int]int, 0)
	m_pMux.streamRecords = make(map[int]int, 0)
	m_pMux.streamTree = make(map[int]int, 0)
	m_pMux.dataStructs = make(map[int]model.DataStruct, 0)
	m_pMux.patVersion = -1
	m_pMux.pmtVersions = make(map[int]int, 0)
	m_pMux.scte35SplicePTS = make(map[int][]int, 0)
}

// Handle incoming data from demuxer
func (m_pMux *tsDemuxPipe) processUnit(buf []byte, pktCnt int) error {
	pkt, tsErr := model.TsPacket(buf)

	if tsErr != nil {
		return tsErr
	}
	buf = pkt.GetPayload()

	tsc, fieldErr := pkt.GetField("tsc")
	if fieldErr != nil {
		return fieldErr
	}

	// If scrambled, throw away
	if tsc != 0 {
		return errors.New("the packet is scrambled")
	}

	// Determine the type of the unit
	pid, fieldErr := pkt.GetField("pid")
	if fieldErr != nil {
		return fieldErr
	}

	pusiInt, fieldErr := pkt.GetField("pusi")
	if fieldErr != nil {
		return fieldErr
	}
	pusi := pusiInt != 0

	afc, fieldErr := pkt.GetField("afc")
	if fieldErr != nil {
		return fieldErr
	}

	cc, fieldErr := pkt.GetField("cc")
	if fieldErr != nil {
		return fieldErr
	}

	inputMon.checkTsHeader(pid, afc, cc, pktCnt)

	pcr := -1

	if pkt.HasAdaptationField() {
		var err error
		pcr, err = pkt.GetValueFromAdaptationField("pcr")
		if err != nil {
			return err
		}

		spliceCountdown, err := pkt.GetValueFromAdaptationField("spliceCountdown")
		if err != nil {
			return err
		}
		if spliceCountdown != -1 {
			if spliceCountdown >= 128 {
				spliceCountdown -= 256
			}
			m_pMux.logger.Info("[%d] At TS packet #%d, spliceCountdown is %d", pid, pktCnt, spliceCountdown)
		}
	}

	switch pid {
	case 0:
		// PAT
		err := m_pMux.handleData(buf, pid, pusi, pktCnt, -1, -1, pcr)
		if err != nil {
			return err
		}
	case 8191:
		// Null packets
	default:
		if pid < 32 {
			// Special pids
			return nil
		}
		hasKey := false
		for _, pmtPid := range m_pMux.programRecords {
			if pid == pmtPid {
				hasKey = true
				break
			}
		}
		if hasKey {
			// PMT
			err := m_pMux.handleData(buf, pid, pusi, pktCnt, -1, -1, pcr)
			if err != nil {
				return err
			}
		} else {
			// Others
			streamType, isKnownStream := m_pMux.streamRecords[pid]
			progNum := m_pMux.streamTree[pid]

			// Contained in PMT, continue the parsing
			if isKnownStream {
				err := m_pMux.handleData(buf, pid, pusi, pktCnt, progNum, streamType, pcr)
				if err != nil {
					return err
				}
			} else {
				// Not contained in PMT
				return nil
			}
		}
	}
	return nil
}

func (m_pMux *tsDemuxPipe) PsiUpdateFinished(pid int, jsonBytes []byte) {
}

func (m_pMux *tsDemuxPipe) SpliceEventReceived(dpiPid int, spliceCmdTypeStr string, splicePTS []int) {
	if spliceCmdTypeStr == "splice_null" {
		return
	}

	progNum := m_pMux.streamTree[dpiPid]
	if _, hasKey := m_pMux.scte35SplicePTS[progNum]; !hasKey {
		m_pMux.scte35SplicePTS[progNum] = make([]int, 0)
	}
	m_pMux.scte35SplicePTS[progNum] = append(m_pMux.scte35SplicePTS[progNum], splicePTS...)

	m_pMux.logger.Info("Received SCTE-35 %s with splice PTS %v", spliceCmdTypeStr, splicePTS)
}

func (m_pMux *tsDemuxPipe) GetPATVersion() int {
	return m_pMux.patVersion
}

func (m_pMux *tsDemuxPipe) GetPmtVersion(progNum int) int {
	if version, hasKey := m_pMux.pmtVersions[progNum]; hasKey {
		return version
	} else {
		return -1
	}
}

func (m_pMux *tsDemuxPipe) AddProgram(version int, progNum int, pmtPid int) {
	if version != m_pMux.patVersion {
		m_pMux.logger.Info("PAT version updated")
	}
	m_pMux.logger.Info("New program added: %d => %d", progNum, pmtPid)
	m_pMux.programRecords[progNum] = pmtPid

	m_pMux.patVersion = version
}

func (m_pMux *tsDemuxPipe) AddStream(version int, progNum int, streamPid int, streamType int) {
	if oldVersion, hasKey := m_pMux.pmtVersions[progNum]; hasKey && oldVersion != -1 {
		m_pMux.logger.Info("PMT version for program %d updated", progNum)
	}
	m_pMux.pmtVersions[progNum] = version

	// Check if this pid already exists
	if oldType, hasKey := m_pMux.streamRecords[streamPid]; hasKey {
		// Check if stream type of the pid changes
		if oldType != streamType {
			m_pMux.logger.Info("Stream type of stream with pid %d updated: %d => %d", streamPid, oldType, streamType)
		}

		// Check if the stream belongs to another program
		for oldPid, oldProgNum := range m_pMux.streamTree {
			if oldPid == streamPid {
				if oldProgNum != progNum {
					m_pMux.logger.Info("Pid %d parent program updated: %d => %d", streamPid, oldProgNum, progNum)
				}
				break
			}
		}
	} else {
		m_pMux.logger.Info("Add stream with pid %d and type %d to program %d", streamPid, streamType, progNum)
	}

	m_pMux.streamRecords[streamPid] = streamType
	m_pMux.streamTree[streamPid] = progNum
}

func (m_pMux *tsDemuxPipe) PesPacketReady(buf common.CmBuf, pid int) {
	outUnit := common.MakeIOUnit(buf, 1, pid)
	m_pMux.outputQueue = append(m_pMux.outputQueue, outUnit)
	m_pMux.control.outputUnitAdded()
	m_pMux.callback.outputReady()
}

func (m_pMux *tsDemuxPipe) GetPmtPidByProgNum(progNum int) int {
	if pid, hasKey := m_pMux.programRecords[progNum]; hasKey {
		return pid
	}
	return -1
}

func (m_pMux *tsDemuxPipe) handleData(buf []byte, pid int, pusi bool, pktCnt int, progNum int, streamType int, pcr int) error {
	if pcr >= 0 {
		clk := m_pMux.control.updateSrcClk(progNum)
		clk.updatePcrRecord(pcr, pktCnt)
	}

	dataProcessed := true

	if pusi {
		if ds, hasKey := m_pMux.dataStructs[pid]; hasKey {
			parseErr := ds.Process()
			delete(m_pMux.dataStructs, pid)
			if parseErr != nil {
				return parseErr
			}
		}
		var ds model.DataStruct
		var err error
		switch m_pMux._getPktType(pid) {
		case PAT:
			fallthrough
		case PMT:
			fallthrough
		case DATA:
			ds, err = model.PsiTable(m_pMux, pktCnt, pid, buf)
		case VIDEO:
			fallthrough
		case AUDIO:
			ds, err = model.PesPacket(m_pMux, buf, pid, pktCnt, progNum, streamType)
		default:
			err = errors.New(fmt.Sprintf("Unknown stream type for pid %d", pid))
		}
		if err != nil {
			return err
		}
		if ds.Ready() {
			parseErr := ds.Process()
			if parseErr != nil {
				return parseErr
			}
			delete(m_pMux.dataStructs, pid)
		} else {
			m_pMux.dataStructs[pid] = ds
		}
	} else if ds, hasKey := m_pMux.dataStructs[pid]; hasKey {
		ds.Append(buf)
		if ds.Ready() {
			parseErr := ds.Process()
			delete(m_pMux.dataStructs, pid)
			if parseErr != nil {
				return parseErr
			}
		}
	} else {
		dataProcessed = false
	}

	if dataProcessed {
		m_pMux.control.dataParsed(pid)
	}

	return nil
}

// Return type of packets
func (m_pMux *tsDemuxPipe) _getPktType(pid int) PKT_TYPE {
	if pid == 0 {
		return PAT
	}

	// Check if PMT pid
	hasKey := false
	for _, pmtPid := range m_pMux.programRecords {
		if pid == pmtPid {
			hasKey = true
			break
		}
	}
	if hasKey {
		return PMT
	}

	// Check stream type
	streamType, isKnownStream := m_pMux.streamRecords[pid]
	if isKnownStream {
		streamName := m_pMux.control.queryStreamType(streamType)
		splitStreamName := strings.Split(streamName, " ")
		streamTypeName := splitStreamName[len(splitStreamName)-1]
		return PKT_TYPE(streamTypeName)
	}

	return UNKNOWN
}

// Duration is independent of program, so just choose one
func (m_pMux *tsDemuxPipe) getDuration() int {
	firstProgNum := -1
	for progNum := range m_pMux.programRecords {
		firstProgNum = progNum
		break
	}
	clk := m_pMux.control.updateSrcClk(firstProgNum)
	start, _ := clk.requestPcr(-1, 0)
	end, _ := clk.requestPcr(-1, m_pMux.control.getInputCount())
	return end - start
}

func (m_pMux *tsDemuxPipe) getOutputUnit() common.CmUnit {
	outUnit := m_pMux.outputQueue[0]
	if len(m_pMux.outputQueue) == 1 {
		m_pMux.outputQueue = make([]common.CmUnit, 0)
	} else {
		m_pMux.outputQueue = m_pMux.outputQueue[1:]
	}
	return outUnit
}

func getDemuxPipe(callback IDemuxCallback, control *demuxController, name string) tsDemuxPipe {
	rv := tsDemuxPipe{
		callback: callback,
		control: control,
		logger: common.CreateLogger(name),
	}
	rv._setup()
	return rv
}
