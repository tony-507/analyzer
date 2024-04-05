package monitor

import (
	"fmt"
	"strings"
	"sync"
	"time"

	tm "github.com/buger/goterm"
	"github.com/tony-507/analyzers/src/plugins/monitor/colors"
	"github.com/tony-507/analyzers/src/plugins/monitor/impl"
	"github.com/tony-507/analyzers/src/tttKernel"
)

var _SLEEP_DURATION = 5 * time.Second

type monitorStat struct {
	inputIdCnt int
}

type monitor struct {
	isRunning    bool
	showTime     bool
	ctrlChan     chan struct{}
	mtx          sync.Mutex
	wg           sync.WaitGroup
	heading      string
	impl         impl.MonitorImpl
	stat         monitorStat
}

func (m *monitor) setParameter(param *OutputMonitorParam) {
	if param.Redundancy != nil {
		m.impl = impl.GetRedundancyMonitor(param.Redundancy)
		m.showTime = true
	} else {
		m.impl = impl.GetScte35Monitor()
	}
}

func (m *monitor) start() {
	m.isRunning = true
	if m.impl == nil {
		panic("MonitorImpl not set")
	}

	go m.worker()
}

func (m *monitor) stop() {
	m.isRunning = false
	m.wg.Wait()
}

func (m *monitor) feed(unit tttKernel.CmUnit, inputId string) {
	m.mtx.Lock()
	defer m.mtx.Unlock()
	if !m.impl.HasInputId(inputId) {
		for _, field := range m.impl.GetFields() {
			s := fmt.Sprintf("%s_%d", field, m.stat.inputIdCnt)
			m.heading += fmt.Sprintf("|%15s", s)
		}
		m.stat.inputIdCnt++
	}
	m.impl.Feed(unit, inputId)
}

// Worker thread functions
func (m *monitor) worker() {
	m.wg.Add(1)
	defer m.wg.Done()

	for m.isRunning {
		select {
		case <-time.After(_SLEEP_DURATION):
			tm.Clear()
			tm.MoveCursor(1, 1)
			tm.Println(m.getOutputMsg())
			tm.Flush()
		case <-m.ctrlChan:
			break
		}
	}
}

func (m *monitor) getOutputMsg() string {
	lineLen := len(m.heading) + 1
	sepStr := fmt.Sprintf(colors.Reset + "%s\n", strings.Repeat(colors.Reset + "-", lineLen))
	msg := ""
	if m.showTime {
		msg += fmt.Sprintf("Current time: %s\n", time.Now().UTC().Format("15:04:05.000000"))
	}
	msg += sepStr
	msg += m.heading + "|\n"
	msg += sepStr

	m.mtx.Lock()
	defer m.mtx.Unlock()

	for _, body := range m.impl.GetDisplayData() {
		if len(body) == 0 {
			// Hide empty row
			continue
		}
		msg += body + "\n"
		msg += sepStr
	}

	return msg
}

func newMonitor() monitor {
	return monitor{
		isRunning: false,
		showTime: false,
		ctrlChan: make(chan struct{}),
		impl: nil,
		heading: "",
		stat: monitorStat{
			inputIdCnt: 1,
		},
	}
}
