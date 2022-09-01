package worker

import (
	"strings"

	"github.com/tony-507/analyzers/src/avContainer"
	"github.com/tony-507/analyzers/src/common"
	"github.com/tony-507/analyzers/src/ioUtils"
)

// A plugin provides unified interface to perform different functionalities

type Plugin struct {
	Work       interface{}            // The struct that performs the work
	interfaces map[string]interface{} // Store interfaces needed
	Name       string
}

func GetPluginByName(inputName string) Plugin {
	// Deduce the type of plugin by name
	splitName := strings.Split(inputName, "_")
	interfaces := make(map[string]interface{}, 0)
	rv := Plugin{}

	switch splitName[0] {
	case "FileReader":
		work := ioUtils.GetReader()
		interfaces["DeliverUnit"] = work.DeliverUnit
		interfaces["FetchUnit"] = work.FetchUnit
		rv = Plugin{Work: work, Name: inputName, interfaces: interfaces}
	case "FileWriter":
		work := ioUtils.GetFileWriter()
		interfaces["DeliverUnit"] = work.DeliverUnit
		rv = Plugin{Work: work, Name: inputName, interfaces: interfaces}
	case "TsDemuxer":
		work := avContainer.GetTsDemuxer()
		interfaces["DeliverUnit"] = work.DeliverUnit
		interfaces["FetchUnit"] = work.FetchUnit
		rv = Plugin{Work: work, Name: inputName, interfaces: interfaces}
	case "Dummy":
		work := GetDummyPlugin()
		interfaces["DeliverUnit"] = work.DeliverUnit
		interfaces["FetchUnit"] = work.FetchUnit
		rv = Plugin{Work: work, Name: inputName, interfaces: interfaces}
	default:
		panic("Unknown plugin name received")
	}
	return rv
}

func (pn *Plugin) DeliverUnit(unit common.CmUnit) (bool, error) {
	f := pn.interfaces["DeliverUnit"].(func(common.CmUnit) (bool, error))
	return f(unit)
}

func (pn *Plugin) FetchUnit() (common.CmUnit, error) {
	f := pn.interfaces["FetchUnit"].(func() (common.CmUnit, error))
	return f()
}
