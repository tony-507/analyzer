package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/tony-507/analyzers/src/ioUtils"
	"github.com/tony-507/analyzers/src/worker"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Wrong number of arguments")
		fmt.Println("Usage: tsa <file>")
		os.Exit(1)
	}

	ex, _ := os.Executable()
	curDir := filepath.Dir(ex)
	fname := os.Args[1]

	inputParam := ioUtils.IOReaderParam{Fname: fname}
	outputParam := ioUtils.IOWriterParam{OutFolder: curDir + strings.TrimSuffix(fname, filepath.Ext(fname)) + "/"}

	w := worker.GetWorker()

	inputPluginParam := worker.ConstructOverallParam("FileReader_1", inputParam, []string{"TsDemuxer_1"})
	demuxPluginParam := worker.ConstructOverallParam("TsDemuxer_1", nil, []string{"DataHandler_1"})
	dataHandlerParam := worker.ConstructOverallParam("DataHandler_1", nil, []string{"FileWriter_1"})
	outputPluginParam := worker.ConstructOverallParam("FileWriter_1", outputParam, []string{})
	workerParams := []worker.OverallParams{inputPluginParam, demuxPluginParam, dataHandlerParam, outputPluginParam}

	w.StartService(workerParams)
}
