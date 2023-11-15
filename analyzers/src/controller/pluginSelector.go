package controller

import (
	"fmt"
	"strings"

	"github.com/tony-507/analyzers/src/plugins/avContainer/tsdemux"
	"github.com/tony-507/analyzers/src/plugins/dataHandler"
	"github.com/tony-507/analyzers/src/plugins/ioUtils"
	"github.com/tony-507/analyzers/src/plugins/monitor"
	"github.com/tony-507/analyzers/src/tttKernel"
)

func selectPlugin(inputName string) tttKernel.IPlugin {
	// Deduce the type of plugin by name
	splitName := strings.Split(inputName, "_")

	switch splitName[0] {
	case "InputReader":
		return ioUtils.InputReader(inputName)
	case "TsDemuxer":
		return tsdemux.TsDemuxer(inputName)
	case "DataHandler":
		return dataHandler.DataHandlerFactory(inputName)
	case "OutputMonitor":
		return monitor.OutputMonitor(inputName)
	default:
		panic(fmt.Sprintf("Unknown plugin name: %s", inputName))
	}
}
