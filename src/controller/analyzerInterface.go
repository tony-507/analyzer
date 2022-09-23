package controller

import (
	"fmt"
	"os"

	"github.com/tony-507/analyzers/src/ioUtils"
	"github.com/tony-507/analyzers/src/worker"
)

type AnalyzerController struct {
	itf      CtrlInterface
	id       string
	provider worker.Worker
}

func (ctrl *AnalyzerController) buildParamException(name string) {
	fmt.Println("Error in building overallParams, unknown field:", name)
	os.Exit(1)
}

// Main function for building up the worker parameters
func (ctrl *AnalyzerController) buildParams() []worker.OverallParams {
	inputParam := ioUtils.IOReaderParam{Fname: ctrl.itf.SourceSetting.FileInput.Fname}
	outputParam := ioUtils.IOWriterParam{OutFolder: ctrl.itf.OutputSetting.DataOutput.OutDir}

	inputPluginParam := worker.ConstructOverallParam("FileReader_1", inputParam, []string{"TsDemuxer_1"})
	demuxPluginParam := worker.ConstructOverallParam("TsDemuxer_1", nil, []string{"DataHandler_1"})
	dataHandlerParam := worker.ConstructOverallParam("DataHandler_1", nil, []string{"FileWriter_1"})
	outputPluginParam := worker.ConstructOverallParam("FileWriter_1", outputParam, []string{})
	return []worker.OverallParams{inputPluginParam, demuxPluginParam, dataHandlerParam, outputPluginParam}
}

func (ctrl *AnalyzerController) StartApp() {
	workerParams := ctrl.buildParams()
	ctrl.provider.StartService(workerParams)
}

func GetController(itf CtrlInterface) AnalyzerController {
	ctrl := AnalyzerController{}
	ctrl.itf = itf
	ctrl.id = "Default"

	ctrl.provider = worker.GetWorker()

	return ctrl
}
