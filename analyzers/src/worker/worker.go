package worker

import (
	"github.com/tony-507/analyzers/src/common"
	"github.com/tony-507/analyzers/src/logs"
	"github.com/tony-507/analyzers/src/resources"
)

// A worker runs a graph to provide a service
// Assumption: The graph does not contain any cyclic subgraph
type Worker struct {
	logger         logs.Log
	graph          Graph
	resourceLoader resources.ResourceLoader
	isRunning      int
}

// Main function for running a graph
func (w *Worker) RunGraph() {
	// Start all plugins with a for loop
	for _, pg := range w.graph.nodes {
		pg.Work.SetResource(&w.resourceLoader)
		pg.SetParameter(pg.m_parameter)
		pg.StartSequence()
	}
	// Suspend the main thread here to keep the pipeline working
	// Break the loop once all plugins have stopped
	for {
		// Break the service loop if stop signal is set
		if w.isRunning == 0 {
			break
		}
	}
}

// Depth-first search
func (w *Worker) _searchNode(name string, curPos *Plugin) *Plugin {
	var rv *Plugin

	if curPos == nil {
		// Start searching
		for _, root := range w.graph.nodes {
			rv = w._searchNode(name, root)
			if rv != nil {
				return rv
			}
		}
		return nil
	}

	if name == curPos.Name {
		return curPos
	}
	if len(curPos.children) == 0 {
		return nil
	}

	// Recursive search
	for _, child := range curPos.children {
		rv = w._searchNode(name, child)
		if rv != nil {
			return rv
		}
	}
	return nil
}

func (w *Worker) PostRequest(name string, unit common.CmUnit) {
	if unit == nil {
		return
	}

	reqType, isReq := unit.GetField("reqType").(common.WORKER_REQUEST)
	if !isReq {
		panic("Error in worker request handling")
	}

	// Check which node this plugin corresponds to
	node := w._searchNode(name, nil)

	switch reqType {
	case common.FETCH_REQUEST:
		outputUnit := node.FetchUnit()
		for _, child := range node.children {
			child.DeliverUnit(outputUnit)
		}
	case common.DELIVER_REQUEST:
		outputUnit := node.parent[0].FetchUnit()
		node.DeliverUnit(outputUnit)
	case common.EOS_REQUEST:
		w.isRunning -= 1
		// Trigger EndSequence of children nodes
		for _, child := range node.children {
			child.EndSequence()
		}
	case common.RESOURCE_REQUEST:
		query, ok := unit.GetBuf().([]string)
		if !ok || len(query) != 2 {
			panic("Wrong resource query format. Should be array of strings")
		}
		w.resourceLoader.Query(query[0], query[1])
	}

}

func (w *Worker) SetGraph(graph Graph) {
	w.graph = graph
	// Recursively set callback for nodes
	graph.SetCallback(w, nil)
	w.isRunning = len(w.graph.nodes)
}

func GetWorker() Worker {
	w := Worker{isRunning: 0, resourceLoader: resources.CreateResourceLoader(), logger: logs.CreateLogger("Worker")}
	return w
}