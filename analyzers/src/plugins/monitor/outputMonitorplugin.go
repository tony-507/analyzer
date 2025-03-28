package monitor

import (
	"encoding/json"
	"strings"

	"github.com/tony-507/analyzers/src/logging"
	"github.com/tony-507/analyzers/src/tttKernel"
)

type outputMonitorPlugin struct {
	logger   logging.Log
	callback tttKernel.RequestHandler
	monitor  monitor
	name     string
}

func (rm *outputMonitorPlugin) SetCallback(callback tttKernel.RequestHandler) {
	rm.callback = callback
}

func (rm *outputMonitorPlugin) SetParameter(paramStr string) {
	var param OutputMonitorParam
	if err := json.Unmarshal([]byte(paramStr), &param); err != nil {
		panic(err)
	}
	rm.monitor.setParameter(&param)
}

func (rm *outputMonitorPlugin) SetResource(*tttKernel.ResourceLoader) {}

func (rm *outputMonitorPlugin) StartSequence() {
	rm.monitor.start()
}

func (rm *outputMonitorPlugin) DeliverUnit(unit tttKernel.CmUnit, inputId string) {
	rm.monitor.feed(unit, inputId)
}

func (rm *outputMonitorPlugin) DeliverStatus(unit tttKernel.CmUnit) {}

func (rm *outputMonitorPlugin) FetchUnit() tttKernel.CmUnit {
	return nil
}

func (rm *outputMonitorPlugin) EndSequence() {
	rm.monitor.stop()
}

func (rm *outputMonitorPlugin) PrintInfo(sb *strings.Builder) {}

func (rm *outputMonitorPlugin) Name() string {
	return rm.name
}

func OutputMonitor(name string) tttKernel.IPlugin {
	return &outputMonitorPlugin{
		logger: logging.CreateLogger("OutputMonitor"),
		name: name,
		monitor: newMonitor(),
	}
}