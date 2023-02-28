package tsdemux

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/tony-507/analyzers/src/common"
	"github.com/tony-507/analyzers/src/plugins/avContainer/model"
)

type tsDemuxPipe struct {
	logger         common.Log
	control        *demuxController // Controller from demuxer
	demuxedBuffers map[int][]byte   // A map mapping pid to bitstreams
	demuxStartCnt  map[int]int      // A map mapping pid to start packet index of demuxedBuffers[pid]
	outputQueue    []common.CmUnit  // Outputs to other plugins
	content        model.PAT
	programs       map[int]model.PMT // A map from program number to PMT
	isRunning      bool
}

func (m_pMux *tsDemuxPipe) _setup() {
	m_pMux.demuxedBuffers = make(map[int][]byte, 0)
	m_pMux.demuxStartCnt = make(map[int]int, 0)
	m_pMux.content = model.PAT{Version: -1}
	m_pMux.programs = make(map[int]model.PMT, 0)
	m_pMux.isRunning = false
}

// Handle incoming data from demuxer
func (m_pMux *tsDemuxPipe) processUnit(buf []byte, pktCnt int) {
	head := model.ReadTsHeader(buf)
	buf = buf[4:]

	// If scrambled, throw away
	if head.Tsc != 0 {
		return
	}

	inputMon.checkTsHeader(head, pktCnt)

	dataProcessed := true // controller use

	// Determine the type of the unit
	pid := head.Pid
	if pid == 0 {
		// PAT
		m_pMux._handlePsiData(buf, pid, head.Pusi, pktCnt, head.Afc)
	} else if pid < 32 {
		// Special pids
		dataProcessed = false
	} else if pid != 8191 {
		// Skip null packet
		_, hasKey := m_pMux.content.ProgramMap[pid]
		if hasKey {
			// PMT
			m_pMux._handlePsiData(buf, pid, head.Pusi, pktCnt, head.Afc)
		} else {
			// Others
			progNum := -1
			streamIdx := -1
			for idx, program := range m_pMux.programs {
				for sIdx, stream := range program.Streams {
					if stream.StreamPid == pid {
						progNum = idx
						streamIdx = sIdx
					}
				}
			}

			// Contained in PMT, continue the parsing
			if progNum != -1 && streamIdx != -1 {
				pktType := m_pMux._getPktType(pid)
				// Determine stream type from last word
				actualTypeSlice := strings.Split(pktType, " ")
				actualType := actualTypeSlice[len(actualTypeSlice)-1]
				switch actualType {
				case "video":
					m_pMux._handleStreamData(buf, pid,
						progNum, head.Pusi, head.Afc,
						pktCnt, int(m_pMux.programs[progNum].Streams[streamIdx].StreamType))
				case "audio":
					m_pMux._handleStreamData(buf, pid,
						progNum, head.Pusi, head.Afc,
						pktCnt, int(m_pMux.programs[progNum].Streams[streamIdx].StreamType))
				case "data":
					m_pMux._handlePsiData(buf, pid, head.Pusi, pktCnt, head.Afc)
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

// Handle PSI data
// Currently only support PAT, PMT and SCTE-35
func (m_pMux *tsDemuxPipe) _handlePsiData(buf []byte, pid int, pusi bool, pktCnt int, afc int) {
	if pusi {
		if len(m_pMux.demuxedBuffers[pid]) != 0 {
			m_pMux._parsePSI(pid, m_pMux.demuxStartCnt[pid], afc)
		} else {
			// Check if we need to update PSI by checking version
			dType := m_pMux._getPktType(pid)
			newVersion := GetVersion(buf)
			switch dType {
			case "PAT":
				if m_pMux.content.Version == -1 {
					m_pMux.demuxedBuffers[pid] = buf
					m_pMux.demuxStartCnt[pid] = pktCnt
					if model.PATReadyForParse(buf) {
						m_pMux._parsePSI(pid, m_pMux.demuxStartCnt[pid], afc)
					}
				} else if m_pMux.content.Version != newVersion {
					m_pMux.logger.Info("PAT version change %d -> %d", m_pMux.content.Version, newVersion)
				}
			case "PMT":
				if len(m_pMux.programs) != 0 {
					progNum := -1
					for idx, program := range m_pMux.programs {
						if program.PmtPid == pid {
							progNum = idx
							break
						}
					}

					if progNum == -1 {
						m_pMux.demuxedBuffers[pid] = buf
						m_pMux.demuxStartCnt[pid] = pktCnt

						if model.PMTReadyForParse(m_pMux.demuxedBuffers[pid]) {
							m_pMux._parsePSI(pid, m_pMux.demuxStartCnt[pid], afc)
						}
					} else if m_pMux.programs[progNum].Version != newVersion {
						m_pMux.logger.Info("PMT at pid %d version change %d -> %d", pid, m_pMux.programs[progNum].Version, newVersion)
					}
				} else {
					m_pMux.demuxedBuffers[pid] = buf
					m_pMux.demuxStartCnt[pid] = pktCnt

					if model.PMTReadyForParse(m_pMux.demuxedBuffers[pid]) {
						m_pMux._parsePSI(pid, m_pMux.demuxStartCnt[pid], afc)
					}
				}
			case "SCTE-35 DPI data":
				m_pMux.demuxedBuffers[pid] = buf
				m_pMux.demuxStartCnt[pid] = pktCnt

				if model.SCTE35ReadyForParse(buf, afc) {
					m_pMux._parsePSI(pid, m_pMux.demuxStartCnt[pid], afc)
				}
			default:
				m_pMux.logger.Error("Don't know how to handle %s", dType)
				panic("What?!")
			}
		}
	} else if len(m_pMux.demuxedBuffers[pid]) != 0 {
		m_pMux.demuxedBuffers[pid] = append(m_pMux.demuxedBuffers[pid], buf...)
	}
}

func (m_pMux *tsDemuxPipe) _parsePSI(pid int, pktCnt int, afc int) {
	// Parse a given buffer
	pktType := m_pMux._getPktType(pid)

	// Output unit related
	var jsonBytes []byte

	switch pktType {
	case "PAT":
		content, err := model.ParsePAT(m_pMux.demuxedBuffers[pid], pktCnt)
		if err != nil {
			m_pMux.control.throwError(pid, pktCnt, err.Error())
		}
		if m_pMux.content.Version == -1 {
			m_pMux.control.updatePidStatus(0, true, 2)
		}
		// Check if PMT pids are updated
		newProgramMap := content.ProgramMap
		for oldPid := range m_pMux.content.ProgramMap {
			hasPid := false
			for newPid := range content.ProgramMap {
				if oldPid == newPid {
					hasPid = true
					newProgramMap[newPid] = -1
					break
				}
			}
			if !hasPid {
				m_pMux.control.updatePidStatus(oldPid, false, 2)
			}
		}
		for pid, progNum := range newProgramMap {
			if progNum > 0 {
				m_pMux.control.updatePidStatus(pid, true, 2)
			}
		}
		m_pMux.content = content
		jsonBytes, _ = json.MarshalIndent(m_pMux.content, "\t", "\t") // Extra tab prefix to support array of Jsons
		m_pMux.logger.Info("PAT parsed: %s", content.ToString())
	case "PMT":
		pmt := model.ParsePMT(m_pMux.demuxedBuffers[pid], pid, pktCnt)

		// Information about parsed PMT
		statMsg := fmt.Sprintf("\nAt pkt#%d\n", pktCnt)
		statMsg += fmt.Sprintf("Program %d\n", pmt.ProgNum)
		for idx, stream := range pmt.Streams {
			statMsg += fmt.Sprintf(" Stream %d: type %s with pid %d\n", idx, m_pMux.control.queryStreamType(stream.StreamType), stream.StreamPid)
		}
		m_pMux.logger.Info(statMsg)

		// Check if PMT pids are updated
		newPmt := pmt
		if oldPmt, hasPmt := m_pMux.programs[pmt.ProgNum]; hasPmt {
			for _, stream := range oldPmt.Streams {
				hasPid := false
				for idx, newStream := range pmt.Streams {
					if stream.StreamPid == newStream.StreamPid {
						hasPid = true
						newPmt.Streams[idx].StreamPid = -1
					}
				}
				if !hasPid {
					m_pMux.control.updatePidStatus(stream.StreamPid, false, 1)
				}
			}
		}
		for _, stream := range newPmt.Streams {
			if stream.StreamPid != -1 {
				m_pMux.control.updatePidStatus(stream.StreamPid, true, 1)
			}
		}

		m_pMux.programs[pmt.ProgNum] = pmt
		jsonBytes, _ = json.MarshalIndent(pmt, "\t", "\t")
	case "SCTE-35 DPI data":
		section := model.ReadSCTE35Section(m_pMux.demuxedBuffers[pid], afc)
		jsonBytes, _ = json.MarshalIndent(section, "\t", "\t")
	default:
		panic("Unknown pid")
	}

	// Clear old buf
	m_pMux.demuxedBuffers[pid] = make([]byte, 0)

	outBuf := common.MakeSimpleBuf(jsonBytes)
	outBuf.SetField("dataType", m_pMux._getPktType(pid), true)

	outUnit := common.MakeIOUnit(outBuf, 2, pid)
	m_pMux.outputQueue = append(m_pMux.outputQueue, outUnit)
}

// Handle stream data
func (m_pMux *tsDemuxPipe) _handleStreamData(buf []byte, pid int, progNum int, pusi bool, afc int, pktCnt int, streamType int) {
	clk := m_pMux.control.updateSrcClk(progNum)

	if afc > 1 {
		af := model.ParseAdaptationField(buf)
		if af.Pcr >= 0 {
			clk.updatePcrRecord(af.Pcr, pktCnt)
		}
		buf = buf[(af.AfLen + 1):]
	}

	// Payload
	if pusi {
		// Initialization issue
		if len(m_pMux.demuxedBuffers[pid]) != 0 {
			pesHeader, headerLen, err := model.ParsePESHeader(m_pMux.demuxedBuffers[pid])
			if err != nil {
				m_pMux.control.throwError(pid, pktCnt, err.Error())
			}

			outBuf := common.MakeSimpleBuf(m_pMux.demuxedBuffers[pid][headerLen:])
			outBuf.SetField("pktCnt", m_pMux.demuxStartCnt[pid], false)
			outBuf.SetField("progNum", progNum, true)
			outBuf.SetField("streamType", streamType, true)
			outBuf.SetField("size", pesHeader.GetSectionLength(), false)
			outBuf.SetField("PTS", pesHeader.GetPts(), false)
			outBuf.SetField("DTS", pesHeader.GetDts(), false)
			outBuf.SetField("dataType", m_pMux._getPktType(pid), true)
			outUnit := common.MakeIOUnit(outBuf, 1, pid)
			m_pMux.outputQueue = append(m_pMux.outputQueue, outUnit)

			m_pMux.demuxedBuffers[pid] = make([]byte, 0)
		}
		m_pMux.demuxedBuffers[pid] = buf
		m_pMux.demuxStartCnt[pid] = pktCnt
	} else if len(m_pMux.demuxedBuffers[pid]) != 0 {
		m_pMux.demuxedBuffers[pid] = append(m_pMux.demuxedBuffers[pid], buf...)
	}
}

func (m_pMux *tsDemuxPipe) _getPktType(pid int) string {
	if pid == 0 {
		return "PAT"
	}

	// Check if PMT pid
	_, hasKey := m_pMux.content.ProgramMap[pid]
	if hasKey {
		return "PMT"
	}

	// Check stream type
	for _, program := range m_pMux.programs {
		for _, stream := range program.Streams {
			if stream.StreamPid == pid {
				return m_pMux.control.queryStreamType(stream.StreamType)
			}
		}
	}

	return ""
}

// Duration is independent of program, so just choose one
func (m_pMux *tsDemuxPipe) getDuration() int {
	firstProgNum := -1
	for _, prog := range m_pMux.programs {
		firstProgNum = prog.ProgNum
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
	return len(m_pMux.programs) > 0 && len(m_pMux.outputQueue) > 0
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
