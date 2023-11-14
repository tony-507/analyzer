package monitor

import (
	"strings"

	"github.com/tony-507/analyzers/src/common"
)

type outputMonitorPlugin struct {
	logger   common.Log
	callback common.RequestHandler
	monitor  monitor
	name     string
}

func (rm *outputMonitorPlugin) SetCallback(callback common.RequestHandler) {
	rm.callback = callback
}

func (rm *outputMonitorPlugin) SetParameter(string) {}

func (rm *outputMonitorPlugin) SetResource(*common.ResourceLoader) {}

func (rm *outputMonitorPlugin) StartSequence() {
	rm.monitor.start()
}

func (rm *outputMonitorPlugin) DeliverUnit(unit common.CmUnit, inputId string) {
	rm.monitor.feed(unit, inputId)
}

func (rm *outputMonitorPlugin) DeliverStatus(unit common.CmUnit) {}

func (rm *outputMonitorPlugin) FetchUnit() common.CmUnit {
	return nil
}

func (rm *outputMonitorPlugin) EndSequence() {
	rm.monitor.stop()
}

func (rm *outputMonitorPlugin) PrintInfo(sb *strings.Builder) {}

func (rm *outputMonitorPlugin) Name() string {
	return rm.name
}

func OutputMonitor(name string) common.IPlugin {
	return &outputMonitorPlugin{
		logger: common.CreateLogger("OutputMonitor"),
		name: name,
		monitor: newMonitor(),
	}
}