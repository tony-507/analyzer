package avContainer

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/tony-507/analyzers/src/common"
)

type demuxEvent struct {
	eventId ctrl_ID
	event   interface{}
}

type tsDemuxPipe struct {
	demuxedBuffers map[int][]byte // A map mapping pid to bitstreams
	demuxStartCnt  map[int]int    // A map mapping pid to start packet index of demuxedBuffers[pid]
	callback       *TsDemuxer
	content        PAT
	programs       []PMT
	isRunning      bool
}

func (m_pMux *tsDemuxPipe) _setup() {
	m_pMux.demuxedBuffers = make(map[int][]byte, 0)
	m_pMux.demuxStartCnt = make(map[int]int, 0)
	m_pMux.content = PAT{Version: -1}
	m_pMux.programs = make([]PMT, 0)
	m_pMux.isRunning = false
}

// Handle incoming data from demuxer
func (m_pMux *tsDemuxPipe) handleUnit(buf []byte, head TsHeader, pktCnt int) {
	// If scrambled, throw away
	if head.tsc != 0 {
		return
	}

	// Determine the type of the unit
	pid := head.pid
	if pid == 0 {
		// PAT
		m_pMux._handlePsiData(buf, pid, head.pusi, pktCnt)
	} else if pid < 32 {
		// Special pids
	} else {
		_, hasKey := m_pMux.content.ProgramMap[pid]
		if hasKey {
			// PMT
			m_pMux._handlePsiData(buf, pid, head.pusi, pktCnt)
		} else {
			// Others
			progIdx := -1
			streamIdx := -1
			for idx, program := range m_pMux.programs {
				for sIdx, stream := range program.Streams {
					if stream.StreamPid == pid {
						progIdx = idx
						streamIdx = sIdx
					}
				}
			}

			// Contained in PMT, continue the parsing
			if progIdx != -1 && streamIdx != -1 {
				pktType := m_pMux._getPktType(pid)
				// Determine stream type from last word
				actualTypeSlice := strings.Split(pktType, " ")
				actualType := actualTypeSlice[len(actualTypeSlice)-1]
				switch actualType {
				case "video":
					m_pMux._handleStreamData(buf, pid,
						m_pMux.programs[progIdx].ProgNum, head.pusi, head.afc,
						pktCnt, int(m_pMux.programs[progIdx].Streams[streamIdx].StreamType))
				case "audio":
					m_pMux._handleStreamData(buf, pid,
						m_pMux.programs[progIdx].ProgNum, head.pusi, head.afc,
						pktCnt, int(m_pMux.programs[progIdx].Streams[streamIdx].StreamType))
				case "data":
					m_pMux._handlePsiData(buf, pid, head.pusi, pktCnt)
				default:
					// Not sure, passthrough first
				}
			}
		}
	}

	m_pMux._postEvent(pid, pktCnt, demuxEvent{eventId: ctrl_PARSEOK, event: pid})
}

// Handle PSI data
// Currently only support PAT and PMT
func (m_pMux *tsDemuxPipe) _handlePsiData(buf []byte, pid int, pusi bool, pktCnt int) {
	// Packet count
	psiBufUnit := common.MakePsiBuf(pktCnt, pid)
	outUnit := common.IOUnit{Buf: psiBufUnit, IoType: 1, Id: -1}
	m_pMux.callback.outputQueue = append(m_pMux.callback.outputQueue, outUnit)

	if pusi {
		if len(m_pMux.demuxedBuffers[pid]) != 0 {
			m_pMux._parsePSI(pid, m_pMux.demuxStartCnt[pid])
		} else {
			// Check if we need to update PSI by checking version
			dType := m_pMux._getPktType(pid)
			newVersion := GetVersion(buf)
			switch dType {
			case "PAT":
				if m_pMux.content.Version == -1 {
					m_pMux.demuxedBuffers[0] = buf
					if PATReadyForParse(buf) {
						m_pMux._parsePSI(pid, m_pMux.demuxStartCnt[pid])
					}
				} else if m_pMux.content.Version != newVersion {
					outMsg := fmt.Sprintf("PAT version change %d -> %d", m_pMux.content.Version, newVersion)
					fmt.Println(outMsg)
				}
			case "PMT":
				if len(m_pMux.programs) != 0 {
					progIdx := -1
					for idx, program := range m_pMux.programs {
						if program.PmtPid == pid {
							progIdx = idx
							break
						}
					}

					if progIdx == -1 {
						m_pMux.demuxedBuffers[pid] = buf
					} else if m_pMux.programs[progIdx].Version != newVersion {
						outMsg := fmt.Sprintf("PMT at pid %d version change %d -> %d", pid, m_pMux.programs[progIdx].Version, newVersion)
						fmt.Println(outMsg)
					}
				} else {
					m_pMux.demuxedBuffers[pid] = buf
					m_pMux.demuxStartCnt[pid] = pktCnt
				}
			default:
				fmt.Println("Don't know how to handle", dType)
				panic("What?!")
			}
		}
	} else if len(m_pMux.demuxedBuffers[pid]) != 0 {
		fmt.Println("Appending...")
		m_pMux.demuxedBuffers[pid] = append(m_pMux.demuxedBuffers[pid], buf...)
	}
}

func (m_pMux *tsDemuxPipe) _parsePSI(pid int, pktCnt int) {
	// Parse a given buffer
	pktType := m_pMux._getPktType(pid)

	// Output unit related
	var outBuf []byte

	switch pktType {
	case "PAT":
		content, err := ParsePAT(m_pMux.demuxedBuffers[pid], pktCnt)
		if err != nil {
			m_pMux._postEvent(pid, pktCnt, err)
		}
		m_pMux.content = content
		outBuf, _ = json.MarshalIndent(m_pMux.content, "\t", "\t") // Extra tab prefix to support array of Jsons
	case "PMT":
		pmt := ParsePMT(m_pMux.demuxedBuffers[pid], pid, pktCnt)

		// Information about parsed PMT
		fmt.Println("\nAt pkt#", pktCnt)
		fmt.Println("Program", pmt.ProgNum)
		for idx, stream := range pmt.Streams {
			fmt.Println(" Stream", idx, ": type", m_pMux._queryStreamType(stream.StreamType), " with pid", stream.StreamPid)
		}

		m_pMux.programs = append(m_pMux.programs, pmt)
		outBuf, _ = json.MarshalIndent(pmt, "\t", "\t")
	default:
		panic("Unknown pid")
	}

	// Clear old buf
	m_pMux.demuxedBuffers[pid] = make([]byte, 0)

	outUnit := common.IOUnit{Buf: outBuf, IoType: 2, Id: pid}
	m_pMux.callback.outputQueue = append(m_pMux.callback.outputQueue, outUnit)
}

// Handle stream data
func (m_pMux *tsDemuxPipe) _handleStreamData(buf []byte, pid int, progNum int, pusi bool, afc int, pktCnt int, streamType int) {
	// Packet count
	psiBufUnit := common.MakePsiBuf(pktCnt, pid)
	outUnit := common.IOUnit{Buf: psiBufUnit, IoType: 1, Id: -1}
	m_pMux.callback.outputQueue = append(m_pMux.callback.outputQueue, outUnit)

	ur := common.GetBufferReader(buf)
	clk := m_pMux.callback._updateSrcClk(progNum)

	if afc > 1 {
		af := ParseAdaptationField(&ur)
		clk.updatePcrRecord(af.pcr, pktCnt)
		buf = ur.GetRemainedBuffer()
	}
	ur = common.GetBufferReader(m_pMux.demuxedBuffers[pid])

	// Payload
	if pusi {
		if len(m_pMux.demuxedBuffers[pid]) != 0 {
			pesHeader, err := ParsePESHeader(ur)
			if err != nil {
				m_pMux._postEvent(pid, m_pMux.demuxStartCnt[pid], err)
			}

			outBuf := common.MakePesBuf(pktCnt, progNum, pesHeader.sectionLen, pesHeader.optionalHeader.pts, pesHeader.optionalHeader.dts, m_pMux.demuxedBuffers[pid], streamType)
			outUnit := common.IOUnit{Buf: outBuf, IoType: 1, Id: pid}
			m_pMux.callback.outputQueue = append(m_pMux.callback.outputQueue, outUnit)

			m_pMux.demuxedBuffers[pid] = make([]byte, 0)
		} else {
			m_pMux.demuxedBuffers[pid] = buf
			m_pMux.demuxStartCnt[pid] = pktCnt
		}
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
				return m_pMux._queryStreamType(stream.StreamType)
			}
		}
	}

	return ""
}

func (m_pMux *tsDemuxPipe) _queryStreamType(typeNum int) string {
	return m_pMux.callback.resourceLoader.Query("streamType", typeNum)
}

// An internal API to post event to demuxer
func (m_pMux *tsDemuxPipe) _postEvent(pid int, pktCnt int, event interface{}) {
	err, isErr := event.(error)
	if isErr {
		eventLevel := ctrl_ERROR
		eventId := ctrl_PARSINGERR
		m_pMux.callback.sendStatus(eventLevel, pid, eventId, pktCnt, err)
	}

	// If not error, it's info
	dEvent, _ := event.(demuxEvent)
	eventLevel := ctrl_INFO
	m_pMux.callback.sendStatus(eventLevel, pid, dEvent.eventId, pktCnt, dEvent.event)
}

// Duration is independent of program, so just choose the first one
func (m_pMux *tsDemuxPipe) getDuration() int {
	clk := m_pMux.callback._updateSrcClk(m_pMux.programs[0].ProgNum)
	start, _ := clk.requestPcr(-1, 0)
	end, _ := clk.requestPcr(-1, m_pMux.callback.pktCnt)
	return end - start
}

func getDemuxPipe(callback *TsDemuxer) tsDemuxPipe {
	rv := tsDemuxPipe{callback: callback}
	rv._setup()
	return rv
}
