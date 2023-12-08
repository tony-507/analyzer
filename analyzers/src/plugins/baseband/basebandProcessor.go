package baseband

import (
	"encoding/json"
	"strings"

	"github.com/tony-507/analyzers/src/common"
	"github.com/tony-507/analyzers/src/common/logging"
	"github.com/tony-507/analyzers/src/plugins/baseband/def"
	"github.com/tony-507/analyzers/src/plugins/baseband/st2110"
	"github.com/tony-507/analyzers/src/tttKernel"
)

type BasebandProcessorPlugin struct {
	logger   logging.Log
	callback tttKernel.RequestHandler
	core     def.ProcessorCore
	name     string
}

func (bb *BasebandProcessorPlugin) SetCallback(callback tttKernel.RequestHandler) {
	bb.callback = callback
}

func (bb *BasebandProcessorPlugin) SetResource(loader *tttKernel.ResourceLoader) {}

func (bb *BasebandProcessorPlugin) SetParameter(param string) {
	var bbParam def.BasebandProcessorParam
	if err := json.Unmarshal([]byte(param), &bbParam); err != nil {
		panic(err)
	}
	bb.core = st2110.St2110ProcessorCore()
}

func (bb *BasebandProcessorPlugin) StartSequence() {
	bb.core.SetCallback(bb)
}

func (bb *BasebandProcessorPlugin) EndSequence() {}

func (bb *BasebandProcessorPlugin) DeliverStatus(status common.CmUnit) {}

func (bb *BasebandProcessorPlugin) DeliverUnit(unit common.CmUnit, inputId string) {
	bb.core.Feed(unit, inputId)
}

func (bb *BasebandProcessorPlugin) FetchUnit() common.CmUnit {
	return nil
}

func (bb *BasebandProcessorPlugin) PrintInfo(sb *strings.Builder) {
	bb.core.PrintInfo(sb)
}

func (bb *BasebandProcessorPlugin) Name() string {
	return bb.name
}

func (bb *BasebandProcessorPlugin) OnDataReady(unit common.CmUnit) {}

func BasebandProcessor(name string) tttKernel.IPlugin {
	return &BasebandProcessorPlugin{
		logger: logging.CreateLogger(name),
		core:   nil,
		name:   name,
	}
}
