package controller

// import (
// 	"os"

// 	"github.com/tony-507/analyzers/src/avContainer/tsdemux"
// 	"github.com/tony-507/analyzers/src/ioUtils"
// 	"github.com/tony-507/analyzers/src/logs"
// 	"github.com/tony-507/analyzers/src/worker"
// )

// type AnalyzerController struct {
// 	logger   logs.Log
// 	itf      CtrlInterface
// 	id       string
// 	provider worker.Worker
// }

// func (ctrl *AnalyzerController) buildParamException(name string) {
// 	ctrl.logger.Log(logs.FATAL, "Error in building overallParams, unknown field: %s", name)
// 	os.Exit(1)
// }

// // Main function for building up the worker parameters
// func (ctrl *AnalyzerController) buildParams() []worker.OverallParams {
// 	rv := []worker.OverallParams{}

// 	inputParam := ioUtils.IOReaderParam{Source: ioUtils.SOURCE_FILE,
// 		FileInput: ioUtils.FileInputParam{Fname: ctrl.itf.SourceSetting.FileInput.Fname}, SkipCnt: ctrl.itf.SourceSetting.SkipCnt,
// 		MaxInCnt: ctrl.itf.SourceSetting.MaxInCnt}
// 	outputParam := ioUtils.IOWriterParam{OutputType: ioUtils.OUTPUT_FILE, FileOutput: ioUtils.FileOutputParam{OutFolder: ctrl.itf.OutputSetting.DataOutput.OutDir}}

// 	// If demuxer is not needed, go to writer directly
// 	if ctrl.itf.DemuxSetting.Mode != 0 {
// 		rv = append(rv, worker.ConstructOverallParam("FileReader_1", inputParam, []string{"TsDemuxer_1"}))

// 		demuxParam := tsdemux.DemuxParams{Mode: tsdemux.DEMUX_FULL}

// 		rv = append(rv, worker.ConstructOverallParam("TsDemuxer_1", demuxParam, []string{"DataHandler_1"}))
// 		rv = append(rv, worker.ConstructOverallParam("DataHandler_1", nil, []string{"FileWriter_1"}))
// 	} else {
// 		rv = append(rv, worker.ConstructOverallParam("FileReader_1", inputParam, []string{"FileWriter_1"}))
// 	}
// 	rv = append(rv, worker.ConstructOverallParam("FileWriter_1", outputParam, []string{}))
// 	return rv
// }

// func (ctrl *AnalyzerController) StartApp() {
// 	workerParams := ctrl.buildParams()
// 	ctrl.provider.StartService(workerParams)
// }

// func GetController(itf CtrlInterface) AnalyzerController {
// 	ctrl := AnalyzerController{logger: logs.CreateLogger("Controller")}
// 	ctrl.itf = itf
// 	ctrl.id = "Default"

// 	ctrl.provider = worker.GetWorker()
// 	ctrl.logger.Log(logs.INFO, "Controller input: %v", itf)

// 	return ctrl
// }
