package tttKernel

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/tony-507/analyzers/src/logging"
)

type workerRequest struct {
	source  string
	reqType WORKER_REQUEST
	body    interface{}
}

// A worker runs a graph to provide a service
// Assumption: The graph does not contain any cyclic subgraph
type Worker struct {
	logger         logging.Log
	isRunning      int
	routineChan    chan struct{}
	nodes          []*graphNode
	reqChannel     chan workerRequest
	resourceLoader ResourceLoader
	statusStore    map[int][]string // Map from msgId to an array of plugin names
	wg             sync.WaitGroup
}

/* APIs for worker */

// Start ttt service
func (w *Worker) StartService(params []OverallParams, selectPlugin func(string) IPlugin) {
	w.setGraph(buildGraph(params, selectPlugin))

	go w.runGraph()

	for w.isRunning != 0 {}

	w.wg.Wait()
}

func (w *Worker) UpdateResource(resource Resource) {
	w.resourceLoader.resource = resource
}

func (w *Worker) StopGraph() {
	if w.isRunning != 0 {
		w.logger.Error("Force stop worker due to unexpected exception")
		for _, pg := range w.nodes {
			pg.stopPlugin()
		}
	}
	w.routineChan <- struct{}{}

	w.printInfo()
	w.isRunning = 0
}

// Main function for running a graph
func (w *Worker) runGraph() {
	w.wg.Add(1)
	defer w.wg.Done()
	defer w.StopGraph()

	startTime := time.Now()
	if w.resourceLoader.IsRedundancyEnabled {
		w.logger.Info("Redundancy monitor is enabled. Output directory %s would be deleted",
			w.resourceLoader.resource.OutDir)
		os.RemoveAll(w.resourceLoader.resource.OutDir)
	}

	for _, node := range w.nodes {
		node.impl.SetParameter(node.m_parameter)
	}
	for _, node := range w.nodes {
		node.impl.SetResource(&w.resourceLoader)
	}
	for _, node := range w.nodes {
		node.impl.StartSequence()
	}
	w.logger.Info("Start up delay: %dms", time.Now().Sub(startTime).Milliseconds())

	go w.startDiagnostics()

	w.handleRequests()
}

// Diagnostics
func (w *Worker) startDiagnostics() {
	w.wg.Add(1)
	defer w.wg.Done()

	for w.isRunning != 0 {
		select {
		case <-time.After(10 * time.Second):
			w.printInfo()
		case <-w.routineChan:
			break
		}
	}
}

func (w *Worker) printInfo() {
	w.logger.Info("Worker\n\tRunning nodes: %d", w.isRunning)
	for _, node := range w.nodes {
		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("Plugin: %s\n", node.name()))
		node.printInfo(&sb)
		w.logger.Info(sb.String())
	}
}

// Callback function
func (w *Worker) onRequestReceived(name string, reqType WORKER_REQUEST, obj interface{}) {
	request := workerRequest{source: name, reqType: reqType, body: obj}
	if w.isRunning != 0 {
		w.reqChannel <- request
	}
}

// Depth-first search
func (w *Worker) searchNode(name string, curPos *graphNode) *graphNode {
	var node *graphNode = nil
	for _, pg := range w.nodes {
		if pg.name() == name {
			node = pg
		}
	}
	return node
}

func (w *Worker) handleRequests() {
	for {
		request := <-w.reqChannel
		w.handleOneRequest(request.source, request.reqType, request.body)
		if w.isRunning == 0 {
			break
		}
	}
}

func (w *Worker) handleOneRequest(name string, reqType WORKER_REQUEST, obj interface{}) {
	switch reqType {
	case POST_REQUEST:
		unit, _ := obj.(CmUnit)
		w.postRequest(name, unit)
	case STATUS_LISTEN_REQUEST:
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
	case STATUS_REQUEST:
		if unit, isValid := obj.(*CmStatusUnit); isValid {
			w.postStatus(unit)
		} else {
			w.logger.Error("Worker error: Receive a status request with invalid unit: %v", obj)
		}
	case ERROR_REQUEST:
		err, _ := obj.(error)
		w.logger.Error("From %s: %s", name, err.Error())
		w.StopGraph()
	default:
		errMsg := fmt.Sprintf("Non-implemented request type %v", reqType)
		w.logger.Error(errMsg)
		panic(errors.New(errMsg))
	}
}

func (w *Worker) postRequest(name string, unit CmUnit) {
	if unit == nil {
		return
	}

	reqType, isReq := unit.GetField("reqType").(WORKER_REQUEST)
	if !isReq {
		panic("Error in worker request handling")
	}

	// Check which node this plugin corresponds to
	node := w.searchNode(name, nil)
	if node == nil {
		w.logger.Error("Received POST request from unknown node %s", name)
	}

	switch reqType {
	case FETCH_REQUEST:
		outputUnit := node.fetchUnit()
		for _, child := range node.children {
			child.deliverUnit(outputUnit, node.name())
		}
	case DELIVER_REQUEST:
		node.deliverUnit(nil, "worker")
	case EOS_REQUEST:
		w.isRunning -= 1
		w.logger.Trace("Worker receives EOS from %s", node.name())
		// Trigger EndSequence of children nodes
		node.stopPlugin()
	}

}

func (w *Worker) postStatus(unit CmUnit) {
	if id, isInt := unit.GetField("id").(int); isInt {
		if arr, hasKey := w.statusStore[id]; hasKey {
			for _, name := range arr {
				node := w.searchNode(name, nil)
				if node == nil {
					w.logger.Error("Fail to deliver status to unknown node %s", name)
				}
				w.logger.Info("Deliver a status to %s", node.name())
				node.deliverStatus(unit)
			}
		}
	}
}

func (w *Worker) setGraph(nodeList []*graphNode) {
	w.nodes = nodeList
	// Recursively set callback for nodes
	for _, node := range w.nodes {
		node.impl.SetCallback(w.onRequestReceived)
		if strings.Contains(node.name(), "Monitor") {
			w.resourceLoader.IsRedundancyEnabled = true
		}
	}

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

func NewWorker() Worker {
	return Worker{
		logger:         logging.CreateLogger("Worker"),
		isRunning:      0,
		routineChan:    make(chan struct{}),
		reqChannel:     make(chan workerRequest, 50),
		resourceLoader: CreateResourceLoader(),
		statusStore:    make(map[int][]string, 0),
	}
}
