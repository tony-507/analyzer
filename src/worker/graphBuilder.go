package worker

import "fmt"

type OverallParams struct {
	pluginName  string
	pluginParam interface{}
	children    []string
}

func ConstructOverallParam(name string, params interface{}, children []string) OverallParams {
	return OverallParams{pluginName: name, pluginParam: params, children: children}
}

// Construct graph from parameters and run the graph
func (w *Worker) StartService(params []OverallParams) {
	graph := GetEmptyGraph()

	createdPlugin := make([]*Plugin, 0)
	// Constrution of graph

	fmt.Println("Start building graph")

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
		}

		pg.setParameterStr(param.pluginParam)

		if pg.isRoot() {
			graph.AddRoot(pg)
		}

		createdPlugin = append(createdPlugin, pg)

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
				AddPath(pg, []*Plugin{pg_new})
			}
		}
	}

	fmt.Println("Start running graph")

	w.SetGraph(graph)
	w.RunGraph()
}
