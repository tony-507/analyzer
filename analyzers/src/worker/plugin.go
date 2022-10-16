package worker

import (
	"strings"
	"fmt"

	"github.com/tony-507/analyzers/src/common"
	"github.com/tony-507/analyzers/src/resources"
)

// A basePlugin provides unified interface to perform different functionalities
type basePlugin interface {
	SetParameter(interface{}) // Set up parameters of the plugin
	SetResource(*resources.ResourceLoader) // Set resources of the plugin, which is loaded from the one of worker
	DeliverUnit(common.CmUnit) (bool, error) // Worker sends a unit to the plugin
	FetchUnit() common.CmUnit // Worker gets a unit from the plugin, and sends it to next plugin(s)
	SetCallback(*Worker) // Set worker callback to allow worker to allocate work
	StartSequence() // Start a plugin. This is called in main thread, so do not suspend in this function
	EndSequence() // Stop a plugin
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
		isRoot := 1
		if splitName[1] == "root" {
			isRoot = 0
		}
		work := GetDummyPlugin(inputName, isRoot)
		rv = initPlugin(&work, inputName, false)
	default:
		msg := fmt.Sprintf("Unknown plugin name: %s", inputName)
		panic(msg)
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
func (pn *Plugin) StartSequence() {
	pn.Work.StartSequence()
}

func (pn *Plugin) EndSequence() {
	pn.Work.EndSequence()
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
