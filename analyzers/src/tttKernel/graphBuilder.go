package tttKernel

import (
	"fmt"

	"github.com/tony-507/analyzers/src/logging"
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

func ConstructOverallParam(name string, params string, children []string) OverallParams {
	return OverallParams{pluginName: name, pluginParam: params, children: children}
}

// Construct graph from parameters and run the graph
func buildGraph(params []OverallParams, selectPlugin func(string) IPlugin) []*graphNode {
	logger := logging.CreateLogger("Worker")
	nodeList := []*graphNode{}

	createdPlugin := make([]*graphNode, 0)
	// Constrution of graph
	paramStr := ""
	for _, param := range params {
		paramStr += param.toString()
	}
	logger.Trace("%s", paramStr)

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
			tmp := getPluginByName(param.pluginName, selectPlugin)
			pg = tmp

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
				tmp := getPluginByName(childName, selectPlugin)
				createdPlugin = append(createdPlugin, tmp)
				nodeList = append(nodeList, tmp)
				addPath(pg, []*graphNode{tmp})
			}
		}
	}

	return nodeList
}
