package tttKernel

import (
	"fmt"
	"strings"

	"github.com/tony-507/analyzers/src/common"
	"github.com/tony-507/analyzers/src/plugins/avContainer/tsdemux"
	"github.com/tony-507/analyzers/src/plugins/dataHandler"
	"github.com/tony-507/analyzers/src/plugins/ioUtils"
)

// A plugin serves as a graph node of operation graph
type _PLUGIN_STATE int

const (
	RUNNING _PLUGIN_STATE = 0
	STOPPED _PLUGIN_STATE = 1
)

type graphNode struct {
	impl        common.IPlugin
	m_parameter string // Store plugin parameters
	m_state     _PLUGIN_STATE
	children    []*graphNode
	parent      []*graphNode
}

// Graph node control flow
func (node *graphNode) stopPlugin() {
	if node.m_state == STOPPED {
		return
	}
	for _, child := range node.children {
		child.impl.EndSequence()
	}
}

// Graph construction
func addPath(parent *graphNode, children []*graphNode) {
	parent.children = append(parent.children, children...)
	for _, child := range children {
		child.parent = append(child.parent, parent)
	}
}

func getPluginByName(inputName string) *graphNode {
	// Deduce the type of plugin by name
	splitName := strings.Split(inputName, "_")
	rv := &graphNode{
		children: make([]*graphNode, 0),
		parent: make([]*graphNode, 0),
		m_state: RUNNING,
	}
	var impl common.IPlugin

	switch splitName[0] {
	case "InputReader":
		impl = ioUtils.InputReader(inputName)
	case "TsDemuxer":
		impl = tsdemux.TsDemuxer(inputName)
	case "DataHandler":
		impl = dataHandler.DataHandlerFactory(inputName)
	case "Dummy":
		role := 1
		if splitName[1] == "root" {
			role = 0
		}
		impl = Dummy(inputName, role)
	default:
		msg := fmt.Sprintf("Unknown plugin name: %s", inputName)
		panic(msg)
	}

	rv.impl = impl

	return rv
}
