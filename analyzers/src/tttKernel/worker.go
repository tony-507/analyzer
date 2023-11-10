package tttKernel

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/tony-507/analyzers/src/common"
)

type workerRequest struct {
	source  string
	reqType common.WORKER_REQUEST
	body    interface{}
}

// A worker runs a graph to provide a service
// Assumption: The graph does not contain any cyclic subgraph
type Worker struct {
	logger         common.Log
	nodes          []*graphNode
	resourceLoader common.ResourceLoader
	isRunning      int
	statusStore    map[int][]string // Map from msgId to an array of plugin names
	reqChannel     chan workerRequest
	wg             sync.WaitGroup
}

// Start ttt service
func (w *Worker) startService(params []OverallParams) {
	w.setGraph(buildGraph(params))

	go w.runGraph()

	for w.isRunning != 0 {}

	w.wg.Wait()
}

// Main function for running a graph
func (w *Worker) runGraph() {
	w.wg.Add(1)
	defer w.wg.Done()

	startTime := time.Now()
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

	timer := time.NewTimer(1 * time.Second)
	for w.isRunning != 0 {
		go func() {
			<-timer.C
			w.printInfo()
			timer.Reset(10 * time.Second)
		}()
	}
	timer.Stop()
	w.printInfo()
}

func (w *Worker) printInfo() {
	for _, node := range w.nodes {
		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("Plugin: %s\n", node.name()))
		node.printInfo(&sb)
		w.logger.Info(sb.String())
	}
}

// Callback function
func (w *Worker) onRequestReceived(name string, reqType common.WORKER_REQUEST, obj interface{}) {
	request := workerRequest{source: name, reqType: reqType, body: obj}
	w.reqChannel <- request
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

	if name == curPos.name() {
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

func (w *Worker) handleRequests() {
	for {
		request := <-w.reqChannel
		w.handleOneRequest(request.source, request.reqType, request.body)
		if w.isRunning == 0 {
			break
		}
	}
}

func (w *Worker) handleOneRequest(name string, reqType common.WORKER_REQUEST, obj interface{}) {
	switch reqType {
	case common.POST_REQUEST:
		unit, _ := obj.(common.CmUnit)
		w.postRequest(name, unit)
	case common.STATUS_LISTEN_REQUEST:
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
	case common.STATUS_REQUEST:
		if unit, isValid := obj.(*common.CmStatusUnit); isValid {
			w.postStatus(unit)
		} else {
			w.logger.Error("Worker error: Receive a status request with invalid unit: %v", obj)
		}
	case common.ERROR_REQUEST:
		err, _ := obj.(error)
		w.logger.Error("From %s: %s", name, err.Error())
		for _, node := range w.nodes {
			node.stopPlugin()
		}
	default:
		errMsg := fmt.Sprintf("Non-implemented request type %v", reqType)
		w.logger.Error(errMsg)
		panic(errors.New(errMsg))
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
		outputUnit := node.fetchUnit()
		for _, child := range node.children {
			child.deliverUnit(outputUnit)
		}
	case common.DELIVER_REQUEST:
		node.deliverUnit(nil)
	case common.EOS_REQUEST:
		w.isRunning -= 1
		w.logger.Trace("Worker receives EOS from %s", node.name())
		// Trigger EndSequence of children nodes
		node.stopPlugin()
	}

}

func (w *Worker) postStatus(unit common.CmUnit) {
	if id, isInt := unit.GetField("id").(int); isInt {
		if arr, hasKey := w.statusStore[id]; hasKey {
			for _, name := range arr {
				node := w.searchNode(name, nil)
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

func getWorker() Worker {
	return Worker{
		isRunning:      0,
		resourceLoader: common.CreateResourceLoader(),
		logger:         common.CreateLogger("Worker"),
		statusStore:    make(map[int][]string, 0),
		reqChannel:     make(chan workerRequest, 50),
	}
}
