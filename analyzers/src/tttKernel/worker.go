package tttKernel

import (
	"fmt"

	"github.com/tony-507/analyzers/src/common"
)

// A worker runs a graph to provide a service
// Assumption: The graph does not contain any cyclic subgraph
type Worker struct {
	logger         common.Log
	nodes          []*graphNode
	resourceLoader common.ResourceLoader
	isRunning      int
	ready          bool             // All plugins are initialized
	statusStore    map[int][]string // Map from msgId to an array of plugin names
	statusList     []common.CmUnit  // Store status from plugins sent before finishing initialization
}

// Main function for running a graph
func (w *Worker) runGraph() {
	var rootPg *graphNode
	// Start all plugins with a for loop
	for _, pg := range w.nodes {
		// Handle root separately to prevent race condition
		if pg.isRoot() {
			rootPg = pg
		}
		pg.impl.SetParameter(pg.m_parameter)
		pg.impl.SetResource(&w.resourceLoader)
		pg.impl.StartSequence()
	}
	w.ready = true
	w.handlePendingStatus()
	rootPg.impl.DeliverUnit(nil)

	// Wait until all plugins stop
	for w.isRunning != 0 {
	}
}

func (w *Worker) handlePendingStatus() {
	for _, unit := range w.statusList {
		w.postStatus(unit)
	}
	w.statusList = make([]common.CmUnit, 0)
}

// Depth-first search
func (w *Worker) searchNode(name string, curPos *graphNode) *graphNode {
	var rv *graphNode

	if curPos == nil {
		// Start searching
		for _, root := range w.nodes {
			rv = w.searchNode(name, root)
			if rv != nil {
				return rv
			}
		}
		return nil
	}

	if name == curPos.impl.Name() {
		return curPos
	}
	if len(curPos.children) == 0 {
		return nil
	}

	// Recursive search
	for _, child := range curPos.children {
		rv = w.searchNode(name, child)
		if rv != nil {
			return rv
		}
	}
	return nil
}

func (w *Worker) handleRequests(name string, reqType common.WORKER_REQUEST, obj interface{}) {
	if reqType == common.POST_REQUEST {
		unit, _ := obj.(common.CmUnit)
		w.postRequest(name, unit)
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
		if unit, isValid := obj.(*common.CmStatusUnit); isValid {
			if w.ready {
				w.postStatus(unit)
			} else {
				w.statusList = append(w.statusList, unit)
			}
		} else {
			w.logger.Error("Worker error: Receive a status request with invalid unit: %v", obj)
		}
	} else if reqType == common.ERROR_REQUEST {
		err, _ := obj.(error)
		w.logger.Error(name, "throws an error")
		panic(err)
	}
}

func (w *Worker) postRequest(name string, unit common.CmUnit) {
	if unit == nil {
		return
	}

	reqType, isReq := unit.GetField("reqType").(common.WORKER_REQUEST)
	if !isReq {
		panic("Error in worker request handling")
	}

	// Check which node this plugin corresponds to
	node := w.searchNode(name, nil)

	switch reqType {
	case common.FETCH_REQUEST:
		outputUnit := node.impl.FetchUnit()
		for _, child := range node.children {
			child.impl.DeliverUnit(outputUnit)
		}
	case common.DELIVER_REQUEST:
		outputUnit := node.parent[0].impl.FetchUnit()
		node.impl.DeliverUnit(outputUnit)
	case common.EOS_REQUEST:
		w.isRunning -= 1
		w.logger.Trace("Worker receives EOS from %s", node.impl.Name())
		// Trigger EndSequence of children nodes
		for _, child := range node.children {
			child.impl.EndSequence()
		}
	case common.RESOURCE_REQUEST:
		query, ok := unit.GetBuf().([]string)
		if !ok || len(query) != 2 {
			panic("Wrong resource query format. Should be array of strings")
		}
		w.resourceLoader.Query(query[0], query[1])
	}

}

func (w *Worker) postStatus(unit common.CmUnit) {
	if id, isInt := unit.GetField("id").(int); isInt {
		if arr, hasKey := w.statusStore[id]; hasKey {
			for _, name := range arr {
				node := w.searchNode(name, nil)
				node.impl.DeliverStatus(unit)
			}
		}
	}
}

func (w *Worker) setGraph(nodeList []*graphNode) {
	w.nodes = nodeList
	// Recursively set callback for nodes
	w.setCallbackForNodes(nil)

	isRunning := 0
	for _, node := range w.nodes {
		isRunning += len(node.children)
		if len(node.parent) == 0 {
			// Root node
			isRunning += 1
		}
	}
	w.isRunning = isRunning
}

func (w *Worker) setCallbackForNodes(curNode *graphNode) {
	if curNode == nil {
		for _, root := range w.nodes {
			w.setCallbackForNodes(root)
		}
	} else {
		if len(curNode.children) != 0 {
			for _, child := range curNode.children {
				w.setCallbackForNodes(child)
			}
		}
		curNode.impl.SetCallback(w.handleRequests)
	}
}

func getWorker() Worker {
	w := Worker{
		isRunning:      0,
		ready:          false,
		resourceLoader: common.CreateResourceLoader(),
		logger:         common.CreateLogger("Worker"),
		statusStore:    make(map[int][]string, 0),
		statusList:     make([]common.CmUnit, 0),
	}
	return w
}
