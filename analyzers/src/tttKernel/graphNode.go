package tttKernel

import (
	"strings"
	"sync"
)

// A plugin serves as a graph node of operation graph
type _PLUGIN_STATE int

const (
	RUNNING _PLUGIN_STATE = 0
	STOPPED _PLUGIN_STATE = 1
)

type graphNode struct {
	impl        IPlugin
	m_parameter string // Store plugin parameters
	m_state     _PLUGIN_STATE
	children    []*graphNode
	parent      []*graphNode
	mtx         sync.Mutex
}

// Graph node control flow
func (node *graphNode) stopPlugin() {
	if node.m_state == STOPPED {
		return
	} else {
		node.mtx.Lock()
		defer node.mtx.Unlock()
		node.impl.EndSequence()
		node.m_state = STOPPED
	}

	for _, child := range node.children {
		child.impl.EndSequence()
	}
}

func (node *graphNode) printInfo(sb *strings.Builder) {
	node.mtx.Lock()
	defer node.mtx.Unlock()
	node.impl.PrintInfo(sb)
}

func (node *graphNode) deliverUnit(unit CmUnit, inputId string) {
	node.mtx.Lock()
	defer node.mtx.Unlock()
	node.impl.DeliverUnit(unit, inputId)
}

func (node *graphNode) fetchUnit() CmUnit {
	node.mtx.Lock()
	defer node.mtx.Unlock()
	return node.impl.FetchUnit()
}

func (node *graphNode) deliverStatus(status CmUnit) {
	node.mtx.Lock()
	defer node.mtx.Unlock()
	node.impl.DeliverStatus(status)
}

func (node *graphNode) name() string {
	node.mtx.Lock()
	defer node.mtx.Unlock()
	return node.impl.Name()
}

// Graph construction
func addPath(parent *graphNode, children []*graphNode) {
	parent.children = append(parent.children, children...)
	for _, child := range children {
		child.parent = append(child.parent, parent)
	}
}

func getPluginByName(inputName string, selectPlugin func(string) IPlugin) *graphNode {
	// Deduce the type of plugin by name
	return &graphNode{
		children: make([]*graphNode, 0),
		parent: make([]*graphNode, 0),
		m_state: RUNNING,
		impl: selectPlugin(inputName),
	}
}
