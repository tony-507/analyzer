package tsdemux

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/tony-507/analyzers/src/common"
	"github.com/tony-507/analyzers/src/common/logging"
	"github.com/tony-507/analyzers/src/plugins/avContainer/model"
	"github.com/tony-507/analyzers/src/tttKernel"
	"github.com/tony-507/analyzers/src/utils"
)

type IDemuxCallback interface {
	outputReady()
	getOutDir() string
}

type tsDemuxPipe struct {
	logger          logging.Log
	callback        IDemuxCallback
	control         *demuxController // Controller from demuxer
	inputMon        inputMonitor
	dataStructs     map[int]model.DataStruct
	fileWriters     map[string]map[int]utils.FileWriter
	programRecords  map[int]int // PAT
	streamRecords   map[int]int // Stream pid => stream type
	streamTree      map[int]int // Stream pid => program number
	patVersion      int
	pmtVersions     map[int]int     // Program number => version
	outputQueue     []tttKernel.CmUnit // Outputs to other plugins
	videoPlayTime   map[int]int     // Program number => playtime
}

func (m_pMux *tsDemuxPipe) _setup() {
	m_pMux.programRecords = make(map[int]int, 0)
	m_pMux.streamRecords = make(map[int]int, 0)
	m_pMux.streamTree = make(map[int]int, 0)
	m_pMux.dataStructs = make(map[int]model.DataStruct, 0)
	m_pMux.patVersion = -1
	m_pMux.pmtVersions = make(map[int]int, 0)
}

func (m_pMux *tsDemuxPipe) start() {}

func (m_pMux *tsDemuxPipe) stop() {
	for fileType, writers := range m_pMux.fileWriters {
		for pid, writer := range writers {
			err := writer.Close()
			if err != nil {
				m_pMux.logger.Error("Fail to close %s writer for pid %d: %s", fileType, pid, err.Error())
			}
		}
	}
}

// Handle incoming data from demuxer
func (m_pMux *tsDemuxPipe) processUnit(buf []byte, pktCnt int) error {
	pkt, tsErr := model.TsPacket(buf)

	if tsErr != nil {
		return tsErr
	}
	buf = pkt.GetPayload()

	// If scrambled, throw away
	if pkt.GetHeader().Tsc != 0 {
		return errors.New("the packet is scrambled")
	}

	// Determine the type of the unit
	pid := pkt.GetHeader().Pid
	pusi := pkt.GetHeader().Pusi
	afc := pkt.GetHeader().Afc
	cc := pkt.GetHeader().Cc

	m_pMux.inputMon.checkTsHeader(pid, afc, cc, pktCnt)

	pcr := -1

	if pkt.HasAdaptationField() {
		pcr = int(pkt.GetAdaptationField().Pcr)

		spliceCountdown := pkt.GetAdaptationField().SpliceCountdown
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
			// Not handled, just record it
			m_pMux.control.dataParsed(pid)
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

func (m_pMux *tsDemuxPipe) PsiUpdateFinished(pid int, version int, jsonBytes []byte) {
	if version == -1 {
		version = 0
		for {
			fname := fmt.Sprintf("%s/%d_%d.json", m_pMux.callback.getOutDir(), pid, version)
			if _, err := os.Stat(fname); errors.Is(err, os.ErrNotExist) {
				break
			}
			version += 1
		}
	}

	writer := utils.RawWriter(m_pMux.callback.getOutDir(), fmt.Sprintf("%d_%d.json", pid, version))
	writer.Open()
	writer.Write(tttKernel.MakeSimpleBuf(jsonBytes))
	writer.Close()
}

func (m_pMux *tsDemuxPipe) SpliceEventReceived(dpiPid int, spliceCmdTypeStr string, splicePTS []int, pktCnt int) {
	if spliceCmdTypeStr == "splice_null" {
		return
	}

	buf := tttKernel.MakeSimpleBuf([]byte{})
	buf.SetField("pktCnt", pktCnt, false)
	buf.SetField("pid", dpiPid, false)
	buf.SetField("streamType", 134, true)

	progNum := m_pMux.streamTree[dpiPid]
	if curVideoPlayTime, ok := m_pMux.videoPlayTime[progNum]; ok {
		for _, spliceTime := range splicePTS {
			preroll := 0
			// Not immediate
			if spliceTime != -1 {
				preroll = spliceTime - curVideoPlayTime
				if preroll < 4 * 90000 {
					m_pMux.logger.Warn("Preroll of SCTE-35 at #%d has short preroll %d", pktCnt, preroll)
				}
			}
			data := common.NewScte35Data(int64(spliceTime), int64(preroll))
			unit := common.NewMediaUnit(buf, common.DATA_UNIT)
			unit.Data = &data
			m_pMux.outputQueue = append(m_pMux.outputQueue, unit)
			m_pMux.control.outputUnitAdded()
			m_pMux.callback.outputReady()
		}
	}
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

func (m_pMux *tsDemuxPipe) PesPacketReady(buf tttKernel.CmBuf, pid int) {
	buf.SetField("pid", pid, true)

	if progNum, ok := tttKernel.GetBufFieldAsInt(buf, "progNum"); ok {
		// Stamp PCR here
		clk := m_pMux.control.updateSrcClk(progNum)

		if curCnt, ok := tttKernel.GetBufFieldAsInt(buf, "pktCnt"); ok {
			pid, _ := tttKernel.GetBufFieldAsInt(buf, "pid")
			pcr, _ := clk.requestPcr(pid, curCnt)
			buf.SetField("pcr", pcr, false)
			if dts, ok := tttKernel.GetBufFieldAsInt(buf, "dts"); ok {
				buf.SetField("delay", dts-pcr/300, false)
			}

			if pts, ok := tttKernel.GetBufFieldAsInt(buf, "pts"); ok {
				m_pMux.videoPlayTime[progNum] = pts
			}

			// Write output
			for _, fileType := range []string{"csv", "pes"} {
				shouldWrite := true

				if _, ok := m_pMux.fileWriters[fileType][pid]; !ok {
					shouldWrite = false
					outDir := m_pMux.callback.getOutDir()
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
					m_pMux.fileWriters[fileType][pid].Write(buf)
				}
			}
		} else {
			m_pMux.logger.Trace("Skip writing PES data due to unknown packet count")
		}
	} else {
		m_pMux.logger.Trace("Skip writing PES data due to unknown program number")
	}

	outUnit := common.NewMediaUnit(buf, common.UNKNOWN_UNIT)
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

func (m_pMux *tsDemuxPipe) getOutputUnit() tttKernel.CmUnit {
	outUnit := m_pMux.outputQueue[0]
	if len(m_pMux.outputQueue) == 1 {
		m_pMux.outputQueue = make([]tttKernel.CmUnit, 0)
	} else {
		m_pMux.outputQueue = m_pMux.outputQueue[1:]
	}
	return outUnit
}

func getDemuxPipe(callback IDemuxCallback, control *demuxController, name string) tsDemuxPipe {
	rv := tsDemuxPipe{
		callback: callback,
		control: control,
		fileWriters: map[string]map[int]utils.FileWriter{
			"csv": {},
			"pes": {},
		},
		logger: logging.CreateLogger(name),
		inputMon: setupInputMonitor(),
		videoPlayTime: map[int]int{},
	}
	rv._setup()
	return rv
}
