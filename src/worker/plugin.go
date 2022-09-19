package worker

import (
	"strings"

	"github.com/tony-507/analyzers/src/common"
)

// A basePlugin provides unified interface to perform different functionalities
type basePlugin interface {
	SetParameter(interface{})
	DeliverUnit(common.CmUnit) (bool, error)
	FetchUnit() common.CmUnit
	SetCallback(*Worker)
	StopPlugin()
}

// A plugin serves as a graph node of operation graph
type Plugin struct {
	Work        basePlugin  // The struct that performs the work
	m_parameter interface{} // Store plugin parameters
	Name        string
	children    []*Plugin
	parent      []*Plugin
	bIsRoot     bool
}

func initPlugin(work basePlugin, name string, bIsRoot bool) Plugin {
	return Plugin{Work: work, Name: name, bIsRoot: bIsRoot, children: make([]*Plugin, 0), parent: make([]*Plugin, 0)}
}

func GetPluginByName(inputName string) Plugin {
	// Deduce the type of plugin by name
	splitName := strings.Split(inputName, "_")
	rv := Plugin{}

	switch splitName[0] {
	case "FileReader":
		work := GetInputReaderPlugin(inputName)
		rv = initPlugin(&work, inputName, true)
	case "FileWriter":
		work := GetFileWriterPlugin(inputName)
		rv = initPlugin(&work, inputName, false)
	case "TsDemuxer":
		work := GetTsDemuxPlugin(inputName)
		rv = initPlugin(&work, inputName, false)
	case "DataHandler":
		work := GetDataHandlerPlugin(inputName)
		rv = initPlugin(&work, inputName, false)
	case "Dummy":
		work := GetDummyPlugin(inputName)
		rv = initPlugin(&work, inputName, false)
	default:
		panic("Unknown plugin name received")
	}
	return rv
}

// Plugin methods

func (pn *Plugin) isRoot() bool {
	return pn.bIsRoot
}

func (pn *Plugin) setParameterStr(m_parameter interface{}) {
	pn.m_parameter = m_parameter
}

// Plugin unified interfaces
func (pn *Plugin) StopPlugin() {
	pn.Work.StopPlugin()
}

func (pn *Plugin) SetParameter(m_parameter interface{}) {
	pn.Work.SetParameter(m_parameter)
}

func (pn *Plugin) DeliverUnit(unit common.CmUnit) (bool, error) {
	return pn.Work.DeliverUnit(unit)
}

func (pn *Plugin) FetchUnit() common.CmUnit {
	return pn.Work.FetchUnit()
}

func (pn *Plugin) SetCallback(w *Worker) {
	pn.Work.SetCallback(w)
}
