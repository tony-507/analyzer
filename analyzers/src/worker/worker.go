package worker

import (
	"fmt"

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
	statusStore    map[int][]string // Map from msgId to an array of plugin names
}

// Main function for running a graph
func (w *Worker) RunGraph() {
	var rootPg *Plugin
	// Start all plugins with a for loop
	for _, pg := range w.graph.nodes {
		// Handle root separately to prevent race condition
		if pg.isRoot() {
			rootPg = pg
		}
		pg.setParameter(pg.m_parameter)
		pg.setResource(&w.resourceLoader)
		pg.startSequence()
	}
	rootPg.deliverUnit(nil)

	// Wait until all plugins stop
	for w.isRunning != 0 {
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

func (w *Worker) HandleRequests(name string, reqType common.WORKER_REQUEST, obj interface{}) {
	if reqType == common.POST_REQUEST {
		unit, _ := obj.(common.CmUnit)
		w.PostRequest(name, unit)
	} else if reqType == common.STATUS_LISTEN_REQUEST {
		if msgId, isInt := obj.(int); isInt {
			if _, hasKey := w.statusStore[msgId]; hasKey {
				w.statusStore[msgId] = append(w.statusStore[msgId], name)
			} else {
				w.statusStore[msgId] = make([]string, 1)
				w.statusStore[msgId][0] = name
			}
		} else {
			panic(fmt.Sprintf("Attempt to listen to a status message with invalid ID: %v", obj))
		}
	} else if reqType == common.STATUS_REQUEST {
		if unit, isValid := obj.(common.CmStatusUnit); isValid {
			w.PostStatus(unit)
		} else {
			w.logger.Log(logs.ERROR, "Worker error: Receive a status request with invalid unit: %v", obj)
		}
	} else if reqType == common.ERROR_REQUEST {
		err, _ := obj.(error)
		w.logger.Log(logs.ERROR, name, "throws an error")
		panic(err)
	}
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
		outputUnit := node.fetchUnit()
		for _, child := range node.children {
			child.deliverUnit(outputUnit)
		}
	case common.DELIVER_REQUEST:
		outputUnit := node.parent[0].fetchUnit()
		node.deliverUnit(outputUnit)
	case common.EOS_REQUEST:
		w.isRunning -= 1
		w.logger.Log(logs.TRACE, "Worker receives EOS from %s", node.Name)
		// Trigger EndSequence of children nodes
		for _, child := range node.children {
			child.endSequence()
		}
	case common.RESOURCE_REQUEST:
		query, ok := unit.GetBuf().([]string)
		if !ok || len(query) != 2 {
			panic("Wrong resource query format. Should be array of strings")
		}
		w.resourceLoader.Query(query[0], query[1])
	}

}

func (w *Worker) PostStatus(unit common.CmUnit) {
	if id, isInt := unit.GetField("id").(int); isInt {
		if arr, hasKey := w.statusStore[id]; hasKey {
			for _, name := range arr {
				node := w._searchNode(name, nil)
				node.deliverStatus(unit)
			}
		}
	}
}

func (w *Worker) SetGraph(graph Graph) {
	w.graph = graph
	// Recursively set callback for nodes
	graph.SetCallback(w, nil)

	isRunning := 0
	for _, node := range w.graph.nodes {
		isRunning += len(node.children)
		if len(node.parent) == 0 {
			// Root node
			isRunning += 1
		}
	}
	w.isRunning = isRunning
}

func GetWorker() Worker {
	w := Worker{isRunning: 0, resourceLoader: resources.CreateResourceLoader(), logger: logs.CreateLogger("Worker"), statusStore: make(map[int][]string, 0)}
	return w
}
