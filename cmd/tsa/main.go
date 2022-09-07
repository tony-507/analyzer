package main

import (
	"github.com/tony-507/analyzers/src/ioUtils"
	"github.com/tony-507/analyzers/src/worker"
)

func main() {
	inputParam := ioUtils.IOReaderParam{Fname: "D:\\assets\\ASCENT.TS"}
	outputParam := ioUtils.IOWriterParam{OutFolder: "D:\\workspace\\analyzers\\output\\ASCENT\\"}

	w := worker.GetWorker()

	inputPluginParam := worker.ConstructOverallParam("FileReader_1", inputParam, []string{"TsDemuxer_1"})
	demuxPluginParam := worker.ConstructOverallParam("TsDemuxer_1", nil, []string{"FileWriter_1"})
	outputPluginParam := worker.ConstructOverallParam("FileWriter_1", outputParam, []string{})
	workerParams := []worker.OverallParams{inputPluginParam, demuxPluginParam, outputPluginParam}

	w.StartService(workerParams)
}
