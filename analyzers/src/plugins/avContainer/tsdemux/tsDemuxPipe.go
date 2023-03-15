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
	control         *demuxController // Controller from demuxer
	dataStructs     map[int]model.DataStruct
	programRecords  map[int]int // PAT
	streamRecords   map[int]int // Stream pid => stream type
	streamTree      map[int]int // Stream pid => program number
	patVersion      int
	pmtVersions     map[int]int     // Program number => version
	demuxedBuffers  map[int][]byte  // A map mapping pid to bitstreams
	demuxStartCnt   map[int]int     // A map mapping pid to start packet index of demuxedBuffers[pid]
	outputQueue     []common.CmUnit // Outputs to other plugins
	scte35SplicePTS map[int][]int   // Program number => Splice PTS
	isRunning       bool
}

func (m_pMux *tsDemuxPipe) _setup() {
	m_pMux.programRecords = make(map[int]int, 0)
	m_pMux.streamRecords = make(map[int]int, 0)
	m_pMux.streamTree = make(map[int]int, 0)
	m_pMux.dataStructs = make(map[int]model.DataStruct, 0)
	m_pMux.patVersion = -1
	m_pMux.pmtVersions = make(map[int]int, 0)
	m_pMux.demuxedBuffers = make(map[int][]byte, 0)
	m_pMux.demuxStartCnt = make(map[int]int, 0)
	m_pMux.scte35SplicePTS = make(map[int][]int, 0)
	m_pMux.isRunning = false
}

// Handle incoming data from demuxer
func (m_pMux *tsDemuxPipe) processUnit(buf []byte, pktCnt int) {
	pkt, tsErr := model.TsPacket(buf)

	if tsErr != nil {
		panic(tsErr)
	}
	buf = pkt.GetPayload()

	tsc, fieldErr := pkt.GetField("tsc")
	if fieldErr != nil {
		panic(fieldErr)
	}

	// If scrambled, throw away
	if tsc != 0 {
		return
	}

	dataProcessed := true // controller use

	// Determine the type of the unit
	pid, fieldErr := pkt.GetField("pid")
	if fieldErr != nil {
		panic(fieldErr)
	}

	pusiInt, fieldErr := pkt.GetField("pusi")
	if fieldErr != nil {
		panic(fieldErr)
	}
	pusi := pusiInt != 0

	afc, fieldErr := pkt.GetField("afc")
	if fieldErr != nil {
		panic(fieldErr)
	}

	cc, fieldErr := pkt.GetField("cc")
	if fieldErr != nil {
		panic(fieldErr)
	}

	inputMon.checkTsHeader(pid, afc, cc, pktCnt)

	pcr := -1

	if pkt.HasAdaptationField() {
		var err error
		pcr, err = pkt.GetValueFromAdaptationField("pcr")
		if err != nil {
			panic(err)
		}

		spliceCountdown, err := pkt.GetValueFromAdaptationField("spliceCountdown")
		if err != nil {
			panic(err)
		}
		if spliceCountdown != -1 {
			if spliceCountdown >= 128 {
				spliceCountdown -= 256
			}
			m_pMux.logger.Info("[%d] At TS packet #%d, spliceCountdown is %d", pid, pktCnt, spliceCountdown)
		}
	}

	if pid == 0 {
		// PAT
		err := m_pMux._handlePsiData(buf, pid, pusi, pktCnt, afc)
		if err != nil {
			panic(err)
		}
	} else if pid < 32 {
		// Special pids
		dataProcessed = false
	} else if pid != 8191 {
		// Skip null packet
		hasKey := false
		for _, pmtPid := range m_pMux.programRecords {
			if pid == pmtPid {
				hasKey = true
				break
			}
		}
		if hasKey {
			// PMT
			err := m_pMux._handlePsiData(buf, pid, pusi, pktCnt, afc)
			if err != nil {
				panic(err)
			}
		} else {
			// Others
			streamType, isKnownStream := m_pMux.streamRecords[pid]
			progNum := m_pMux.streamTree[pid]

			// Contained in PMT, continue the parsing
			if isKnownStream {
				pktType := m_pMux._getPktType(pid)
				// Determine stream type from last word
				actualTypeSlice := strings.Split(pktType, " ")
				actualType := actualTypeSlice[len(actualTypeSlice)-1]
				switch actualType {
				case "video":
					err := m_pMux._handleStreamData(buf, pid,
						progNum, pusi, pktCnt, streamType, pcr)
					if err != nil {
						panic(err)
					}
				case "audio":
					err := m_pMux._handleStreamData(buf, pid,
						progNum, pusi, pktCnt, streamType, pcr)
					if err != nil {
						panic(err)
					}
				case "data":
					err := m_pMux._handlePsiData(buf, pid, pusi, pktCnt, afc)
					if err != nil {
						panic(err)
					}
				default:
					// Not sure, passthrough first
					dataProcessed = false
				}
			} else {
				// Not contained in PMT
				dataProcessed = false
			}
		}
	}
	if dataProcessed {
		m_pMux.control.dataParsed(pid)
	}
}

func (m_pMux *tsDemuxPipe) PsiUpdateFinished(pid int, jsonBytes []byte) {
	outBuf := common.MakeSimpleBuf(jsonBytes)
	outBuf.SetField("dataType", m_pMux._getPktType(pid), true)
	outBuf.SetField("streamType", -1, true)

	outUnit := common.MakeIOUnit(outBuf, 2, pid)
	m_pMux.outputQueue = append(m_pMux.outputQueue, outUnit)
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
	if oldPmtPid, hasKey := m_pMux.programRecords[progNum]; hasKey {
		m_pMux.logger.Info("PAT version updated")
		if oldPmtPid != pmtPid {
			m_pMux.control.updatePidStatus(oldPmtPid, false, 2)
		}
		// Remove old PMT streams
	}
	m_pMux.control.updatePidStatus(pmtPid, true, 2)
	m_pMux.logger.Info("New program added: %d => %d", progNum, pmtPid)
	m_pMux.programRecords[progNum] = pmtPid

	if m_pMux.patVersion == -1 {
		m_pMux.control.updatePidStatus(0, true, 2)
	}
	m_pMux.patVersion = version
}

func (m_pMux *tsDemuxPipe) AddStream(version int, progNum int, streamPid int, streamType int) {
	if oldVersion, hasKey := m_pMux.pmtVersions[progNum]; hasKey && oldVersion != -1 {
		m_pMux.logger.Info("PMT version for program %d updated", progNum)
	}
	m_pMux.pmtVersions[progNum] = version

	updateStreamStatus := true

	// Check if this pid already exists
	if oldType, hasKey := m_pMux.streamRecords[streamPid]; hasKey {
		// Check if stream type of the pid changes
		if oldType != streamType {
			m_pMux.logger.Info("Stream type of stream with pid %d updated: %d => %d", streamPid, oldType, streamType)
			actualTypeSlice := strings.Split(m_pMux.control.queryStreamType(oldType), " ")
			actualType := actualTypeSlice[len(actualTypeSlice)-1]
			if actualType == "data" {
				m_pMux.control.updatePidStatus(streamPid, false, 2)
			} else {
				m_pMux.control.updatePidStatus(streamPid, false, 1)
			}
		} else {
			updateStreamStatus = true
		}

		// Check if the stream belongs to another program
		for oldPid, oldProgNum := range m_pMux.streamTree {
			if oldPid == streamPid && oldProgNum != progNum {
				m_pMux.logger.Info("Pid %d parent program updated: %d => %d", streamPid, oldProgNum, progNum)
				break
			}
		}
	} else {
		m_pMux.logger.Info("Add stream with pid %d and type %d to program %d", streamPid, streamType, progNum)
	}

	m_pMux.streamRecords[streamPid] = streamType
	m_pMux.streamTree[streamPid] = progNum

	if updateStreamStatus {
		actualTypeSlice := strings.Split(m_pMux.control.queryStreamType(streamType), " ")
		actualType := actualTypeSlice[len(actualTypeSlice)-1]
		if actualType == "data" {
			m_pMux.control.updatePidStatus(streamPid, true, 2)
		} else {
			m_pMux.control.updatePidStatus(streamPid, true, 1)
		}
	}
}

func (m_pMux *tsDemuxPipe) GetPmtPidByProgNum(progNum int) int {
	if pid, hasKey := m_pMux.programRecords[progNum]; hasKey {
		return pid
	}
	return -1
}

func (m_pMux *tsDemuxPipe) _handlePsiData(buf []byte, pid int, pusi bool, pktCnt int, afc int) error {
	if pusi {
		if m_pMux.dataStructs[pid] != nil {
			table, ok := m_pMux.dataStructs[pid].(model.TableStruct)
			if !ok {
				return errors.New(fmt.Sprintf("Not a table at pid %d", pid))
			}
			parseErr := table.ParsePayload()
			return parseErr
		} else {
			table, err := model.PsiTable(m_pMux, pktCnt, pid, buf)
			if err != nil {
				return err
			}
			if table.Ready() {
				parseErr := table.ParsePayload()
				if parseErr != nil {
					return parseErr
				}
			} else {
				m_pMux.dataStructs[pid] = table
			}
		}
	} else if m_pMux.dataStructs[pid] != nil {
		m_pMux.dataStructs[pid].Append(buf)
		if m_pMux.dataStructs[pid].Ready() {
			table, ok := m_pMux.dataStructs[pid].(model.TableStruct)
			if !ok {
				return errors.New(fmt.Sprintf("Not a table at pid %d", pid))
			}
			parseErr := table.ParsePayload()
			if parseErr != nil {
				return parseErr
			}
			m_pMux.dataStructs[pid] = nil
		}
	} else {
		return errors.New("drop table without preceding section data")
	}
	return nil
}

// Handle stream data
func (m_pMux *tsDemuxPipe) _handleStreamData(buf []byte, pid int, progNum int, pusi bool, pktCnt int, streamType int, pcr int) error {
	if pcr >= 0 {
		clk := m_pMux.control.updateSrcClk(progNum)
		clk.updatePcrRecord(pcr, pktCnt)
	}

	if pusi {
		if m_pMux.dataStructs[pid] != nil {
			outBuf := m_pMux.dataStructs[pid].GetHeader()
			outUnit := common.MakeIOUnit(outBuf, 1, pid)
			m_pMux.outputQueue = append(m_pMux.outputQueue, outUnit)
			m_pMux.dataStructs[pid] = nil
		}

		pesPkt, err := model.PesPacket(buf, pktCnt, progNum, streamType)
		if err != nil {
			return err
		}

		if pesPkt.Ready() {
			outBuf := pesPkt.GetHeader()
			outUnit := common.MakeIOUnit(outBuf, 1, pid)
			m_pMux.outputQueue = append(m_pMux.outputQueue, outUnit)
		} else {
			m_pMux.dataStructs[pid] = pesPkt
		}
	} else if m_pMux.dataStructs[pid] != nil {
		m_pMux.dataStructs[pid].Append(buf)
	} else {
		m_pMux.logger.Error("[%d] drop TS packet with pid %d without preceding section data", pktCnt, pid)
	}

	// Payload
	// if pusi {
	// 	// Initialization issue
	// 	if len(m_pMux.demuxedBuffers[pid]) != 0 {
	// 		pesHeader, headerLen, err := model.ParsePESHeader(m_pMux.demuxedBuffers[pid])
	// 		if err != nil {
	// 			m_pMux.control.throwError(pid, m_pMux.demuxStartCnt[pid], err.Error())
	// 		}

	// 		outBuf := common.MakeSimpleBuf(m_pMux.demuxedBuffers[pid][headerLen:])
	// 		outBuf.SetField("pktCnt", m_pMux.demuxStartCnt[pid], false)
	// 		outBuf.SetField("progNum", progNum, true)
	// 		outBuf.SetField("streamType", streamType, true)
	// 		outBuf.SetField("size", pesHeader.GetSectionLength(), false)
	// 		outBuf.SetField("PTS", pesHeader.GetPts(), false)
	// 		outBuf.SetField("DTS", pesHeader.GetDts(), false)
	// 		outBuf.SetField("dataType", m_pMux._getPktType(pid), true)
	// 		outUnit := common.MakeIOUnit(outBuf, 1, pid)
	// 		m_pMux.outputQueue = append(m_pMux.outputQueue, outUnit)

	// 		splicePTSList := m_pMux.scte35SplicePTS[progNum]
	// 		if len(splicePTSList) > 0 && pesHeader.GetPts() == splicePTSList[0] {
	// 			fmt.Println(fmt.Sprintf("[%d] At packet #%d, PTS matches for SCTE-35 splice time %d", pid, m_pMux.demuxStartCnt[pid], splicePTSList[0]))
	// 			if len(splicePTSList) == 1 {
	// 				splicePTSList = make([]int, 0)
	// 			} else {
	// 				splicePTSList = splicePTSList[1:]
	// 			}
	// 			m_pMux.scte35SplicePTS[progNum] = splicePTSList
	// 		}

	// 		m_pMux.demuxedBuffers[pid] = make([]byte, 0)
	// 	}
	// 	m_pMux.demuxedBuffers[pid] = buf
	// 	m_pMux.demuxStartCnt[pid] = pktCnt
	// } else if len(m_pMux.demuxedBuffers[pid]) != 0 {
	// 	m_pMux.demuxedBuffers[pid] = append(m_pMux.demuxedBuffers[pid], buf...)
	// }
	return nil
}

func (m_pMux *tsDemuxPipe) _getPktType(pid int) string {
	if pid == 0 {
		return "PAT"
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
		return "PMT"
	}

	// Check stream type
	streamType, isKnownStream := m_pMux.streamRecords[pid]
	if isKnownStream {
		return m_pMux.control.queryStreamType(streamType)
	}

	return ""
}

// Duration is independent of program, so just choose one
func (m_pMux *tsDemuxPipe) getDuration() int {
	firstProgNum := -1
	for progNum, _ := range m_pMux.programRecords {
		firstProgNum = progNum
		break
	}
	clk := m_pMux.control.updateSrcClk(firstProgNum)
	start, _ := clk.requestPcr(-1, 0)
	end, _ := clk.requestPcr(-1, m_pMux.control.getInputCount())
	m_pMux.logger.Info("Time elapsed: %d, %d", start, end)
	return end - start
}

func (m_pMux *tsDemuxPipe) getProgramNumber(idx int) int {
	return idx
}

func (m_pMux *tsDemuxPipe) readyForFetch() bool {
	return len(m_pMux.streamRecords) > 0 && len(m_pMux.outputQueue) > 0
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

func getDemuxPipe(control *demuxController, name string) tsDemuxPipe {
	rv := tsDemuxPipe{control: control, logger: common.CreateLogger(name)}
	rv._setup()
	return rv
}
