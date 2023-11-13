package dataHandler

import (
	"strings"

	"github.com/tony-507/analyzers/src/common"
	"github.com/tony-507/analyzers/src/plugins/dataHandler/audio"
	"github.com/tony-507/analyzers/src/plugins/dataHandler/utils"
	"github.com/tony-507/analyzers/src/plugins/dataHandler/video"
)

/*
 * Data flow: input -> data handlers -> data processors -> output
 *
 * Data handler handle input based on specs
 * Data processor handles parsed data from data handler
 */

type DataHandlerFactoryPlugin struct {
	logger     common.Log
	callback   common.RequestHandler
	handlers   map[int]utils.DataHandler
	outputUnit []common.CmUnit
	isRunning  bool
	name       string
	processors []utils.DataProcessor
}

func (df *DataHandlerFactoryPlugin) SetCallback(callback common.RequestHandler) {
	df.callback = callback
}

func (df *DataHandlerFactoryPlugin) SetParameter(m_parameter string) {
	df._setup()
}

func (df *DataHandlerFactoryPlugin) SetResource(loader *common.ResourceLoader) {}

func (df *DataHandlerFactoryPlugin) _setup() {
	df.logger = common.CreateLogger(df.name)
	df.handlers = map[int]utils.DataHandler{}
	df.outputUnit = []common.CmUnit{}
	df.isRunning = true
}

func (df *DataHandlerFactoryPlugin) StartSequence() {
	df.logger.Info("Data handler factory is started")
	for _, proc := range df.processors {
		if err := proc.Start(); err != nil {
			df.logger.Warn("Skip processing due to %s", err.Error())
		}
	}
}

func (df *DataHandlerFactoryPlugin) EndSequence() {
	df.logger.Info("Data handler factory is stopped")
	for _, proc := range df.processors {
		proc.Stop()
	}
	eosUnit := common.MakeReqUnit(df.name, common.EOS_REQUEST)
	common.Post_request(df.callback, df.name, eosUnit)
}

func (df *DataHandlerFactoryPlugin) DeliverUnit(unit common.CmUnit, inputId string) {
	if unit == nil {
		return
	}

	// Extract buffer from input unit
	cmBuf := unit.GetBuf()
	pid, isPidInt := unit.GetField("id").(int)
	if !isPidInt {
		panic("Something wrong with the data")
	}

	_, hasPid := df.handlers[pid]
	dType, ok := common.GetBufFieldAsInt(cmBuf, "streamType")
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
		case 135:
			df.handlers[pid] = audio.AC3Handler(pid)
		}
	}
	if h, hasHandle := df.handlers[pid]; hasHandle {
		newData := utils.CreateParsedData()
		h.Feed(unit, &newData)
		for _, proc := range df.processors {
			proc.Process(cmBuf, &newData)
		}
	}

	df.outputUnit = append(df.outputUnit, unit)

	// Directly output the unit
	reqUnit := common.MakeReqUnit(df.name, common.FETCH_REQUEST)
	common.Post_request(df.callback, df.name, reqUnit)
}

func (df *DataHandlerFactoryPlugin) DeliverStatus(unit common.CmUnit) {}

func (df *DataHandlerFactoryPlugin) FetchUnit() common.CmUnit {
	switch len(df.outputUnit) {
	case 0:
		return nil
	case 1:
		rv := df.outputUnit[0]
		df.outputUnit = make([]common.CmUnit, 0)
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

func DataHandlerFactory(name string) common.IPlugin {
	rv := DataHandlerFactoryPlugin{
		name: name,
		processors: []utils.DataProcessor{
			videoDataProcessor(),
		},
	}
	return &rv
}
