package controller

import (
	"fmt"
	"strings"

	"github.com/tony-507/analyzers/src/common"
	"github.com/tony-507/analyzers/src/plugins/avContainer/tsdemux"
	"github.com/tony-507/analyzers/src/plugins/dataHandler"
	"github.com/tony-507/analyzers/src/plugins/ioUtils"
	"github.com/tony-507/analyzers/src/plugins/monitor"
)

func selectPlugin(inputName string) common.IPlugin {
	// Deduce the type of plugin by name
	splitName := strings.Split(inputName, "_")
	var rv common.IPlugin

	switch splitName[0] {
	case "InputReader":
		rv = ioUtils.InputReader(inputName)
	case "TsDemuxer":
		rv = tsdemux.TsDemuxer(inputName)
	case "DataHandler":
		rv = dataHandler.DataHandlerFactory(inputName)
	case "OutputMonitor":
		rv = monitor.OutputMonitor(inputName)
	default:
		msg := fmt.Sprintf("Unknown plugin name: %s", inputName)
		panic(msg)
	}

	return rv
}
