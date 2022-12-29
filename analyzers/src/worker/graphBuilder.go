package worker

import (
	"fmt"

	"github.com/tony-507/analyzers/src/logs"
)

type OverallParams struct {
	pluginName  string
	pluginParam interface{}
	children    []string
}

func (param OverallParams) toString() string {
	str := fmt.Sprintf("Name: %s\n", param.pluginName)
	str += fmt.Sprintf("Parameters: %v\n", param.pluginParam)
	str += fmt.Sprintf("Children: %v\n\n", param.children)
	return str
}

func ConstructOverallParam(name string, params interface{}, children []string) OverallParams {
	return OverallParams{pluginName: name, pluginParam: params, children: children}
}

func (w *Worker) StartService(params []OverallParams) {
	w.SetGraph(buildGraph(params))
	w.RunGraph()
}

// Construct graph from parameters and run the graph
func buildGraph(params []OverallParams) Graph {
	logger := logs.CreateLogger("GraphBuilder")
	graph := GetEmptyGraph()

	createdPlugin := make([]*Plugin, 0)
	// Constrution of graph
	paramStr := ""
	for _, param := range params {
		paramStr += param.toString()
	}
	logger.Log(logs.TRACE, "Start building graph with parameters %s", paramStr)

	/*
		Note that, unlike in C, it's perfectly OK to return the address of a local variable;
		the storage associated with the variable survives after the function returns.
	*/
	for _, param := range params {
		// Create node if not exist
		bExist := false
		var pg *Plugin
		for _, node := range createdPlugin {
			if node.Name == param.pluginName {
				pg = node
				bExist = true
			}
		}
		if !bExist {
			tmp := GetPluginByName(param.pluginName)
			pg = &tmp

			graph.AddNode(pg)
			createdPlugin = append(createdPlugin, pg)
		}

		pg.setParameterStr(param.pluginParam)

		// Handle children nodes
		for _, childName := range param.children {
			bExist = false

			for _, node := range createdPlugin {
				if node.Name == childName {
					AddPath(pg, []*Plugin{node})
					bExist = true
					break
				}
			}
			if !bExist {
				tmp := GetPluginByName(childName)
				pg_new := &tmp
				createdPlugin = append(createdPlugin, pg_new)
				graph.AddNode(pg_new)
				AddPath(pg, []*Plugin{pg_new})
			}
		}
	}

	statMsg := "Start running graph:"
	for _, node := range graph.nodes {
		statMsg += "\n\tName: " + node.Name
		statMsg += fmt.Sprintf("\n\tParameters: %v", node.m_parameter)
		statMsg += fmt.Sprintf("\n\tOutput: %v", node.children)
		statMsg += "\n"
	}
	logger.Log(logs.TRACE, statMsg)

	return graph
}
