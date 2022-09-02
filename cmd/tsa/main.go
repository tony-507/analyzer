package main

import (
	"github.com/tony-507/analyzers/src/ioUtils"
	"github.com/tony-507/analyzers/src/worker"
)

func main() {
	inputParam := ioUtils.IOReaderParam{Fname: "D:\\assets\\ASCENT.TS"}
	outputParam := ioUtils.IOWriterParam{OutFolder: "D:\\workspace\\analyzers\\output\\ASCENT\\"}

	p1 := worker.GetPluginByName("FileReader_1")
	p2 := worker.GetPluginByName("TsDemuxer_1")
	p3 := worker.GetPluginByName("FileWriter_1")

	n1 := worker.CreateNode(&p1, inputParam)
	n2 := worker.CreateNode(&p2, nil)
	n3 := worker.CreateNode(&p3, outputParam)

	g := worker.GetEmptyGraph()
	g.AddRoot(&n1)
	worker.AddPath(&n1, []*worker.GraphNode{&n2})
	worker.AddPath(&n2, []*worker.GraphNode{&n3})

	w := worker.GetWorker()
	w.SetGraph(g)

	w.RunGraph()
}
