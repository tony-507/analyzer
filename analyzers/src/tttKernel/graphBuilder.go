package tttKernel

import (
	"fmt"

	"github.com/tony-507/analyzers/src/common"
)

type OverallParams struct {
	pluginName  string
	pluginParam string
	children    []string
}

func (param OverallParams) toString() string {
	str := fmt.Sprintf("Name: %s\n", param.pluginName)
	str += fmt.Sprintf("Parameters: %v\n", param.pluginParam)
	str += fmt.Sprintf("Children: %v\n\n", param.children)
	return str
}

func constructOverallParam(name string, params string, children []string) OverallParams {
	return OverallParams{pluginName: name, pluginParam: params, children: children}
}

func (w *Worker) startService(params []OverallParams) {
	w.setGraph(buildGraph(params))
	w.runGraph()
}

// Construct graph from parameters and run the graph
func buildGraph(params []OverallParams) []*graphNode {
	logger := common.CreateLogger("Worker")
	nodeList := []*graphNode{}

	createdPlugin := make([]*graphNode, 0)
	// Constrution of graph
	paramStr := ""
	for _, param := range params {
		paramStr += param.toString()
	}

	/*
		Note that, unlike in C, it's perfectly OK to return the address of a local variable;
		the storage associated with the variable survives after the function returns.
	*/
	for _, param := range params {
		// Create node if not exist
		bExist := false
		var pg *graphNode
		for _, node := range createdPlugin {
			if node.impl.Name() == param.pluginName {
				pg = node
				bExist = true
			}
		}
		if !bExist {
			tmp := getPluginByName(param.pluginName)
			pg = &tmp

			nodeList = append(nodeList, pg)
			createdPlugin = append(createdPlugin, pg)
		}

		pg.m_parameter = param.pluginParam

		// Handle children nodes
		for _, childName := range param.children {
			bExist = false

			for _, node := range createdPlugin {
				if node.impl.Name() == childName {
					addPath(pg, []*graphNode{node})
					bExist = true
					break
				}
			}
			if !bExist {
				tmp := getPluginByName(childName)
				pg_new := &tmp
				createdPlugin = append(createdPlugin, pg_new)
				nodeList = append(nodeList, pg_new)
				addPath(pg, []*graphNode{pg_new})
			}
		}
	}

	statMsg := fmt.Sprintf("Start running graph:\n")
	for _, node := range nodeList {
		statMsg += "\n\tName: " + node.impl.Name()
		statMsg += fmt.Sprintf("\n\tParameters: %v", node.m_parameter)
		statMsg += fmt.Sprintf("\n\tOutput: %v", node.children)
		statMsg += "\n"
	}
	logger.Trace(statMsg)

	return nodeList
}
