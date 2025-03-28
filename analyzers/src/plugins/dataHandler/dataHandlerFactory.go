package dataHandler

import (
	"strings"

	"github.com/tony-507/analyzers/src/logging"
	"github.com/tony-507/analyzers/src/plugins/common"
	"github.com/tony-507/analyzers/src/plugins/dataHandler/audio"
	"github.com/tony-507/analyzers/src/plugins/dataHandler/data"
	"github.com/tony-507/analyzers/src/plugins/dataHandler/utils"
	"github.com/tony-507/analyzers/src/plugins/dataHandler/video"
	"github.com/tony-507/analyzers/src/tttKernel"
)

/*
 * Data flow: input -> data handlers -> data processors -> output
 *
 * Data handler handle input based on specs
 * Data processor handles parsed data from data handler
 */

type DataHandlerFactoryPlugin struct {
	logger     logging.Log
	callback   tttKernel.RequestHandler
	handlers   map[int]utils.DataHandler
	outputUnit []tttKernel.CmUnit
	isRunning  bool
	name       string
	processors []utils.DataProcessor
	loader   *tttKernel.ResourceLoader
}

func (df *DataHandlerFactoryPlugin) SetCallback(callback tttKernel.RequestHandler) {
	df.callback = callback
}

func (df *DataHandlerFactoryPlugin) SetParameter(m_parameter string) {
	df._setup()
}

func (df *DataHandlerFactoryPlugin) SetResource(loader *tttKernel.ResourceLoader) {
	df.loader = loader
}

func (df *DataHandlerFactoryPlugin) _setup() {
	df.logger = logging.CreateLogger(df.name)
	df.handlers = map[int]utils.DataHandler{}
	df.outputUnit = []tttKernel.CmUnit{}
	df.isRunning = true
}

func (df *DataHandlerFactoryPlugin) StartSequence() {
	df.processors = append(df.processors, videoDataProcessor(df.loader.Query("outDir", nil)))

	for _, proc := range df.processors {
		if err := proc.Start(); err != nil {
			df.logger.Warn("Skip processing due to %s", err.Error())
		}
	}

	df.logger.Info("Data handler factory is started")
}

func (df *DataHandlerFactoryPlugin) EndSequence() {
	for _, proc := range df.processors {
		proc.Stop()
	}
	eosUnit := tttKernel.MakeReqUnit(df.name, tttKernel.EOS_REQUEST)
	tttKernel.Post_request(df.callback, df.name, eosUnit)
	df.logger.Info("Data handler factory is stopped")
}

func (df *DataHandlerFactoryPlugin) DeliverUnit(unit tttKernel.CmUnit, inputId string) {
	if unit == nil {
		return
	}

	// Extract buffer from input unit
	cmBuf := unit.GetBuf()
	pid, isPidInt := tttKernel.GetBufFieldAsInt(cmBuf, "pid")
	if !isPidInt {
		panic("Something wrong with the data")
	}

	_, hasPid := df.handlers[pid]
	dType, ok := tttKernel.GetBufFieldAsInt(cmBuf, "streamType")
	if !ok {
		return
	}
	if !hasPid {
		switch dType {
		case 2:
			df.handlers[pid] = video.MPEG2VideoHandler(pid)
		case 27:
			df.handlers[pid] = video.H264VideoHandler(pid)
		case 129:
			df.handlers[pid] = audio.AC3Handler(pid)
		case 134:
			df.handlers[pid] = data.Scte35Handler(pid)
		case 135:
			df.handlers[pid] = audio.AC3Handler(pid)
		}
	}

	var newUnit tttKernel.CmUnit = nil

	if h, hasHandle := df.handlers[pid]; hasHandle {
		newData := utils.CreateParsedData()
		h.Feed(unit, &newData)
		switch newData.GetType() {
		case utils.PARSED_VIDEO:
			newUnit = common.NewMediaUnit(cmBuf, common.VIDEO_UNIT)
		case utils.PARSED_DATA:
			newUnit = common.NewMediaUnit(cmBuf, common.DATA_UNIT)
		default:
			newUnit = common.NewMediaUnit(cmBuf, common.UNKNOWN_UNIT)
		}
		for _, proc := range df.processors {
			proc.Process(newUnit, &newData)
		}
	}

	df.outputUnit = append(df.outputUnit, newUnit)

	// Directly output the unit
	reqUnit := tttKernel.MakeReqUnit(df.name, tttKernel.FETCH_REQUEST)
	tttKernel.Post_request(df.callback, df.name, reqUnit)
}

func (df *DataHandlerFactoryPlugin) DeliverStatus(unit tttKernel.CmUnit) {}

func (df *DataHandlerFactoryPlugin) FetchUnit() tttKernel.CmUnit {
	switch len(df.outputUnit) {
	case 0:
		return nil
	case 1:
		rv := df.outputUnit[0]
		df.outputUnit = make([]tttKernel.CmUnit, 0)
		return rv
	default:
		rv := df.outputUnit[0]
		df.outputUnit = df.outputUnit[1:]
		return rv
	}
}
func (df *DataHandlerFactoryPlugin) PrintInfo(sb *strings.Builder) {
	for _, proc := range df.processors {
		proc.PrintInfo(sb)
	}
}

func (df *DataHandlerFactoryPlugin) Name() string {
	return df.name
}

func DataHandlerFactory(name string) tttKernel.IPlugin {
	rv := DataHandlerFactoryPlugin{
		name: name,
		processors: []utils.DataProcessor{},
	}
	return &rv
}
