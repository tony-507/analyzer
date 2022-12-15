package worker

import (
	"fmt"
	"strings"

	"github.com/tony-507/analyzers/src/avContainer/tsdemux"
	"github.com/tony-507/analyzers/src/common"
	datahandler "github.com/tony-507/analyzers/src/dataHandler"
	"github.com/tony-507/analyzers/src/ioUtils"
	"github.com/tony-507/analyzers/src/resources"
)

// A basePlugin provides unified interface to perform different functionalities
type basePlugin interface {
	SetParameter(interface{})                // Set up parameters of the plugin
	SetResource(*resources.ResourceLoader)   // Set resources of the plugin, which is loaded from the one of worker
	DeliverUnit(common.CmUnit) (bool, error) // Worker sends a unit to the plugin
	FetchUnit() common.CmUnit                // Worker gets a unit from the plugin, and sends it to next plugin(s)
	SetCallback(*Worker)                     // Set worker callback to allow worker to allocate work
	StartSequence()                          // Start a plugin. This is called in main thread, so do not suspend in this function
	EndSequence()                            // Stop a plugin
}

// A plugin serves as a graph node of operation graph
type Plugin struct {
	Work          interface{} // The struct that performs the work
	setCallback   func(common.RequestHandler)
	setParameter  func(interface{})
	setResource   func(*resources.ResourceLoader)
	startSequence func()
	deliverUnit   func(common.CmUnit)
	fetchUnit     func() common.CmUnit
	endSequence   func()
	m_parameter   interface{} // Store plugin parameters
	Name          string
	children      []*Plugin
	parent        []*Plugin
	bIsRoot       bool // Specify if a particular plugin is root node
}

func initPlugin(work interface{}, name string, bIsRoot bool) Plugin {
	return Plugin{Work: work, Name: name, bIsRoot: bIsRoot, children: make([]*Plugin, 0), parent: make([]*Plugin, 0)}
}

func GetPluginByName(inputName string) Plugin {
	// Deduce the type of plugin by name
	splitName := strings.Split(inputName, "_")
	rv := Plugin{}

	switch splitName[0] {
	case "FileReader":
		work := ioUtils.GetReader(inputName)
		rv = initPlugin(&work, inputName, true)
		rv.setCallback = work.SetCallback
		rv.setParameter = work.SetParameter
		rv.setResource = work.SetResource
		rv.startSequence = work.StartSequence
		rv.deliverUnit = work.DeliverUnit
		rv.fetchUnit = work.FetchUnit
		rv.endSequence = work.EndSequence
		rv.bIsRoot = true // This plugin must be a root
	case "FileWriter":
		work := ioUtils.GetOutputWriter(inputName)
		rv = initPlugin(&work, inputName, false)
		rv.setCallback = work.SetCallback
		rv.setParameter = work.SetParameter
		rv.setResource = work.SetResource
		rv.startSequence = work.StartSequence
		rv.deliverUnit = work.DeliverUnit
		rv.fetchUnit = work.FetchUnit
		rv.endSequence = work.EndSequence
		rv.bIsRoot = false // This plugin must be a root
	case "TsDemuxer":
		work := tsdemux.GetTsDemuxer(inputName)
		rv = initPlugin(&work, inputName, false)
		rv.setCallback = work.SetCallback
		rv.setParameter = work.SetParameter
		rv.setResource = work.SetResource
		rv.startSequence = work.StartSequence
		rv.deliverUnit = work.DeliverUnit
		rv.fetchUnit = work.FetchUnit
		rv.endSequence = work.EndSequence
		rv.bIsRoot = false // This plugin must be a root
	case "DataHandler":
		work := datahandler.GetDataHandlerFactory(inputName)
		rv = initPlugin(&work, inputName, false)
		rv.setCallback = work.SetCallback
		rv.setParameter = work.SetParameter
		rv.setResource = work.SetResource
		rv.startSequence = work.StartSequence
		rv.deliverUnit = work.DeliverUnit
		rv.fetchUnit = work.FetchUnit
		rv.endSequence = work.EndSequence
		rv.bIsRoot = false // This plugin must be a root
	case "Dummy":
		isRoot := 1
		if splitName[1] == "root" {
			isRoot = 0
		}
		work := GetDummyPlugin(inputName, isRoot)
		rv = initPlugin(&work, inputName, false)
		if isRoot == 0 {
			rv.bIsRoot = true
		} else {
			rv.bIsRoot = false
		}
		rv.setCallback = work.SetCallback
		rv.setParameter = work.SetParameter
		rv.setResource = work.SetResource
		rv.startSequence = work.StartSequence
		rv.deliverUnit = work.DeliverUnit
		rv.fetchUnit = work.FetchUnit
		rv.endSequence = work.EndSequence
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
