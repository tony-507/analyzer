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
type graphNode struct {
	impl        common.Plugin
	m_parameter string // Store plugin parameters
	children    []*graphNode
	parent      []*graphNode
}

func addPath(parent *graphNode, children []*graphNode) {
	parent.children = append(parent.children, children...)
	for _, child := range children {
		child.parent = append(child.parent, parent)
	}
}

func getPluginByName(inputName string) graphNode {
	// Deduce the type of plugin by name
	splitName := strings.Split(inputName, "_")
	rv := graphNode{children: make([]*graphNode, 0), parent: make([]*graphNode, 0)}
	var impl common.Plugin

	switch splitName[0] {
	case "InputReader":
		impl = ioUtils.GetInputReader(inputName)
	case "OutputWriter":
		impl = ioUtils.GetOutputWriter(inputName)
	case "TsDemuxer":
		impl = tsdemux.GetTsDemuxer(inputName)
	case "DataHandler":
		impl = dataHandler.GetDataHandlerFactory(inputName)
	case "Dummy":
		isRoot := 1
		if splitName[1] == "root" {
			isRoot = 0
		}
		impl = getDummyPlugin(inputName, isRoot)
	default:
		msg := fmt.Sprintf("Unknown plugin name: %s", inputName)
		panic(msg)
	}

	rv.impl = impl

	return rv
}

// Plugin methods

func (pn *graphNode) isRoot() bool {
	return pn.impl.IsRoot()
}

func (pn *graphNode) setParameterStr(m_parameter string) {
	pn.m_parameter = m_parameter
}
