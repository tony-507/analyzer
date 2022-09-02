package worker

import (
	"github.com/tony-507/analyzers/src/common"
)

// A worker runs a graph to provide a service
// Assumption: The graph does not contain any cyclic subgraph
type Worker struct {
	graph     Graph
	isRunning int
}

// Main function for running a graph
func (w *Worker) RunGraph() {
	w._setup()
	roots := w.graph.GetRoots()
	dummyInput := common.IOUnit{Buf: 1, IoType: 0, Id: 0}
	for {
		// Break the service loop if stop signal is set
		if w.isRunning == 0 {
			break
		}
		// Keep sending rubbish to root to kickstart the pipeline
		for _, root := range roots {
			root.val.DeliverUnit(dummyInput)
		}
	}
	w._shutDown()
}

// Set up plugins in graph by breadth-first search
func (w *Worker) _setup() {
	nodes := w.graph.roots
	for len(nodes) != 0 {
		tmpList := make([]*GraphNode, 0)
		for _, node := range nodes {
			node.val.SetParameter(node.m_parameter)
			tmpList = append(tmpList, node.children...)
		}
		nodes = append(nodes, tmpList...)
		nodes = nodes[1:]
	}
}

// Shut down plugins in graph by breadth-first search
func (w *Worker) _shutDown() {
	nodes := w.graph.roots
	for len(nodes) != 0 {
		tmpList := make([]*GraphNode, 0)
		for _, node := range nodes {
			node.val.StopPlugin()
			tmpList = append(tmpList, node.children...)
		}
		nodes = append(nodes, tmpList...)
		nodes = nodes[1:]
	}
}

// Depth-first search
func (w *Worker) _searchNode(name string, curPos *GraphNode) *GraphNode {
	var rv *GraphNode

	if curPos == nil {
		// Start searching
		for _, root := range w.graph.roots {
			rv = w._searchNode(name, root)
			if rv != nil {
				return rv
			}
		}
		return nil
	}

	if name == curPos.val.Name {
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
		outputUnit := node.val.FetchUnit()
		for _, child := range node.children {
			child.val.DeliverUnit(outputUnit)
		}
	case common.DELIVER_REQUEST:
		outputUnit := node.parent[0].val.FetchUnit()
		node.val.DeliverUnit(outputUnit)
	case common.EOS_REQUEST:
		w.isRunning -= 1
	}

}

func (w *Worker) SetGraph(graph Graph) {
	w.graph = graph
	// Recursively set callback for nodes
	graph.SetCallback(w, nil)
	w.isRunning = len(w.graph.roots)
}

func GetWorker() Worker {
	w := Worker{isRunning: 0}
	return w
}
