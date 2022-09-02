package worker

import (
	"strings"

	"github.com/tony-507/analyzers/src/common"
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
		work := GetInputReaderPlugin(inputName)
		interfaces["DeliverUnit"] = work.DeliverUnit
		interfaces["FetchUnit"] = work.FetchUnit
		interfaces["SetCallback"] = work.SetCallback
		interfaces["StopPlugin"] = work.StopPlugin
		interfaces["SetParameter"] = work.SetParameter
		rv = Plugin{Work: work, Name: inputName, interfaces: interfaces}
	case "FileWriter":
		work := GetFileWriterPlugin(inputName)
		interfaces["DeliverUnit"] = work.DeliverUnit
		interfaces["FetchUnit"] = work.FetchUnit
		interfaces["SetCallback"] = work.SetCallback
		interfaces["StopPlugin"] = work.StopPlugin
		interfaces["SetParameter"] = work.SetParameter
		rv = Plugin{Work: work, Name: inputName, interfaces: interfaces}
	case "TsDemuxer":
		work := GetTsDemuxPlugin(inputName)
		interfaces["DeliverUnit"] = work.DeliverUnit
		interfaces["FetchUnit"] = work.FetchUnit
		interfaces["SetCallback"] = work.SetCallback
		interfaces["StopPlugin"] = work.StopPlugin
		interfaces["SetParameter"] = work.SetParameter
		rv = Plugin{Work: work, Name: inputName, interfaces: interfaces}
	case "Dummy":
		work := GetDummyPlugin(inputName)
		// work.SetCallback(&rv)
		interfaces["DeliverUnit"] = work.DeliverUnit
		interfaces["FetchUnit"] = work.FetchUnit
		interfaces["SetCallback"] = work.SetCallback
		interfaces["StopPlugin"] = work.StopPlugin
		interfaces["SetParameter"] = work.SetParameter
		rv = Plugin{Work: work, Name: inputName, interfaces: interfaces}
	default:
		panic("Unknown plugin name received")
	}
	return rv
}

// Plugin unified interfaces

func (pn *Plugin) StopPlugin() {
	f := pn.interfaces["StopPlugin"].(func())
	f()
}

func (pn *Plugin) SetParameter(m_parameter interface{}) {
	f := pn.interfaces["SetParameter"].(func(interface{}))
	f(m_parameter)
}

func (pn *Plugin) DeliverUnit(unit common.CmUnit) (bool, error) {
	f := pn.interfaces["DeliverUnit"].(func(common.CmUnit) (bool, error))
	return f(unit)
}

func (pn *Plugin) FetchUnit() common.CmUnit {
	f := pn.interfaces["FetchUnit"].(func() common.CmUnit)
	return f()
}

func (pn *Plugin) SetCallback(w *Worker) {
	f := pn.interfaces["SetCallback"].(func(*Worker))
	f(w)
}
