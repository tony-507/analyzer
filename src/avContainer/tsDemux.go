package avContainer

import (
	"fmt"
	"sync"
	"time"

	"github.com/tony-507/analyzers/src/common"
)

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

type ctrl_LEVEL int

const (
	ctrl_INFO  ctrl_LEVEL = 1
	ctrl_ERROR ctrl_LEVEL = 2
)

type ctrl_ID int

const (
	ctrl_PARSINGERR     ctrl_ID = 1 // Error
	ctrl_PARSEOK        ctrl_ID = 2 // OK
	ctrl_NUMERICAL_RISK ctrl_ID = 3 // Potential computational issue, print message to state that
)

type controlParam struct {
	id     ctrl_ID     // 1: successful packet count, 2: parsing error
	pid    int         // pid of the packet
	curCnt int         // the packet number at which the param is set
	data   interface{} // Depends on id
	level  ctrl_LEVEL
}

type TsDemuxer struct {
	demuxPipe   *tsDemuxPipe           // Actual demuxing operation
	control     chan common.CmUnit     // Controller channel, separate from demuxMap to prevent race condition
	progClkMap  map[int]*programSrcClk // progNum -> srcClk
	outputQueue []common.IOUnit        // Outputs to other plugins
	isRunning   int                    // Counting channels, similar to waitGroup
	pktCnt      int                    // The index of currently fed packet
	wg          sync.WaitGroup
}

func (m_pMux *TsDemuxer) _setup() {
	m_pMux.control = make(chan common.CmUnit)
	m_pMux.progClkMap = make(map[int]*programSrcClk, 0)
	m_pMux.outputQueue = make([]common.IOUnit, 0)
	m_pMux.pktCnt = 1
	m_pMux.isRunning = 0

	pipe := getDemuxPipe(m_pMux)
	m_pMux.demuxPipe = &pipe

	m_pMux.wg.Add(2)
	go m_pMux._setupDemuxControl()
	go m_pMux._setupMonitor()
}

// Demuxer monitor, run as a Goroutine to monitor demuxer's status
// Currently only check if demuxer gets stuck
func (m_pMux *TsDemuxer) _setupMonitor() {
	defer m_pMux.wg.Done()

	assertTimeout := 5

	// Wait for a handler to be created first, so it won't exit on initialization
	for m_pMux.isRunning == 0 {
		continue
	}

	for {
		if m_pMux.isRunning == 0 {
			break
		}

		// Check stuck
		curCnt := m_pMux.pktCnt
		time.Sleep(5 * time.Second)
		if m_pMux.pktCnt == curCnt {
			if m_pMux.isRunning == 0 {
				break
			}
			fmt.Println("\ntsDemuxer status")
			fmt.Println("\tCurrent count:", curCnt)
			fmt.Println("\tisRunning:", m_pMux.isRunning == 1)
			fmt.Println("\tOutput queue size:", len(m_pMux.outputQueue))

			assertTimeout -= 1
		}
	}
}

// Demuxer internal controller, run as a Goroutine to handle control messages
func (m_pMux *TsDemuxer) _setupDemuxControl() {
	// Parsing status related
	pktCnt := make(map[int]int, 0)
	m_pMux.isRunning += 1

	defer m_pMux.wg.Done()
	for {
		unit := <-m_pMux.control
		msgId, _ := unit.GetField("id").(common.CM_STATUS)

		if msgId == common.STATUS_CONTROL_DATA {
			param, _ := unit.GetField("body").(controlParam)
			switch param.level {
			case ctrl_ERROR:
				// Parsing error
				err, _ := param.data.(error)
				outMsg := fmt.Sprintf("[%d] At pkt#%d, %s", param.pid, param.curCnt, err.Error())
				panic(outMsg)
			case ctrl_INFO:
				// Inform changes
				switch param.id {
				case ctrl_PARSEOK:
					pid, _ := param.data.(int)
					pktCnt[pid] += 1
				case ctrl_NUMERICAL_RISK:
					infoMsg, _ := param.data.(string)
					fmt.Println(infoMsg)
				default:
					panic("Unknown control id received at monitor")
				}
			default:
				panic("Unknown control level received at monitor")
			}
		} else if msgId == common.STATUS_END_ROUTINE {
			break
		}
	}

	duration := m_pMux.demuxPipe.getDuration()
	m_pMux.isRunning -= 1
	sum := 0
	rateSum := 0.0
	fmt.Println("\nTS statistics:")
	fmt.Printf("TS duration: %fs\n", float64(duration)/27000000)
	fmt.Println("-------------------------------------------------")
	fmt.Printf("|    pid    |   count   |  bitrate  | frequency |\n")
	fmt.Printf("|-----------|-----------|-----------|-----------|\n")
	for pid, cnt := range pktCnt {
		rate := float64(cnt) * 1504 * 27000000 / float64(duration)
		rateSum += rate
		fmt.Printf("|%11d|%11d|%11.2f|%11.2f|\n", pid, cnt, rate, rate/1504)
		sum += cnt
	}
	fmt.Println("-------------------------------------------------")
	fmt.Printf("|%11s|%11d|%11.2f|%11s|\n", "", sum, rateSum, "")
	fmt.Println("-------------------------------------------------")
}

func (m_pMux *TsDemuxer) sendStatus(level ctrl_LEVEL, pid int, id ctrl_ID, curCnt int, data interface{}) {
	param := controlParam{id: id, pid: pid, curCnt: curCnt, data: data, level: level}
	controlUnit := common.MakeStatusUnit(common.STATUS_CONTROL_DATA, common.STATUS_CONTROL_DATA, param)
	m_pMux.control <- controlUnit
}

func (m_pMux *TsDemuxer) _updateSrcClk(progNum int) *programSrcClk {
	_, hasKey := m_pMux.progClkMap[progNum]
	if !hasKey {
		clk := getProgramSrcClk(m_pMux)
		m_pMux.progClkMap[progNum] = &clk
	}
	return m_pMux.progClkMap[progNum]
}

func (m_pMux *TsDemuxer) StopPlugin() {
	// time.Sleep(3 * time.Second)
	fmt.Println("Shutting down handlers")
	unit := common.MakeStatusUnit(common.STATUS_END_ROUTINE, common.STATUS_END_ROUTINE, "")
	m_pMux.control <- unit
	m_pMux.wg.Wait()
	fmt.Println("Demuxer done")
}

func (m_pMux *TsDemuxer) FetchUnit() common.IOUnit {
	outLen := len(m_pMux.outputQueue)

	if outLen != 0 {
		rv := m_pMux.outputQueue[0]

		pesBuf, isPes := rv.Buf.(common.PesBuf)
		if isPes {
			progNum, _ := pesBuf.GetField("progNum").(int)

			// sanity check
			if progNum == -1 {
				fmt.Println(rv)
				panic("Error at fetchunit")
			}

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
				clk := m_pMux._updateSrcClk(m_pMux.demuxPipe.programs[0].ProgNum)

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
		return rv
	}

	rv := common.IOUnit{IoType: 0}
	return rv
}

func (m_pMux *TsDemuxer) DeliverUnit(inUnit common.CmUnit) {
	// Perform demuxing on the received TS packet
	inBuf, _ := inUnit.GetBuf().([]byte)
	r := common.GetBufferReader(inBuf)
	tsHeader := ReadTsHeader(&r)

	m_pMux.demuxPipe.handleUnit(r.GetRemainedBuffer(), tsHeader, m_pMux.pktCnt)

	m_pMux.pktCnt += 1
}

func GetTsDemuxer() *TsDemuxer {
	m_pMux := TsDemuxer{}
	m_pMux._setup()
	return &m_pMux
}