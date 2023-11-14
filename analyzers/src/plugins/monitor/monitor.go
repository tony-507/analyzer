package monitor

import (
	"fmt"
	"strings"
	"sync"
	"time"

	tm "github.com/buger/goterm"
	"github.com/tony-507/analyzers/src/common"
	"github.com/tony-507/analyzers/src/plugins/monitor/colors"
	"github.com/tony-507/analyzers/src/plugins/monitor/impl"
)

var _SLEEP_DURATION = 5 * time.Second

type monitorStat struct {
	inputIdCnt int
}

type monitor struct {
	isRunning    bool
	displayTimer *time.Timer
	mtx          sync.Mutex
	wg           sync.WaitGroup
	heading      string
	impl         impl.MonitorImpl
	stat         monitorStat
}

func (m *monitor) start() {
	m.isRunning = true
	m.impl = impl.GetRedundancyMonitor()
	m.displayTimer = time.NewTimer(_SLEEP_DURATION)

	go m.worker()
}

func (m *monitor) stop() {
	m.isRunning = false
	m.wg.Wait()
}

func (m *monitor) feed(unit common.CmUnit, inputId string) {
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
		m.displayTimer.Reset(_SLEEP_DURATION)
		<-m.displayTimer.C

		tm.Clear()
		tm.MoveCursor(1, 1)
		tm.Println(m.getOutputMsg())
		tm.Flush()
	}
	m.displayTimer.Stop()
}

func (m *monitor) getOutputMsg() string {
	sepStr := fmt.Sprintf(colors.Reset + "%s\n", strings.Repeat(colors.Reset + "-", len(m.heading) + 1))
	msg := sepStr
	msg += m.heading + "|\n"
	msg += sepStr

	m.mtx.Lock()
	defer m.mtx.Unlock()

	for _, body := range m.impl.GetDisplayData() {
		msg += body + "\n"
		msg += sepStr
	}

	return msg
}

func newMonitor() monitor {
	return monitor{
		isRunning: false,
		displayTimer: nil,
		impl: nil,
		heading: "",
		stat: monitorStat{
			inputIdCnt: 1,
		},
	}
}
