package dataHandler

import (
	"github.com/tony-507/analyzers/src/common"
	"github.com/tony-507/analyzers/src/plugins/dataHandler/audio"
	"github.com/tony-507/analyzers/src/plugins/dataHandler/utils"
	"github.com/tony-507/analyzers/src/plugins/dataHandler/video"
)

type DataHandlerFactoryPlugin struct {
	logger     common.Log
	callback   common.RequestHandler
	handlers   map[int]utils.DataHandler
	outputUnit []common.CmUnit
	isRunning  bool
	name       string
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
}

func (df *DataHandlerFactoryPlugin) EndSequence() {
	df.logger.Info("Data handler factory is stopped")
	eosUnit := common.MakeReqUnit(df.name, common.EOS_REQUEST)
	common.Post_request(df.callback, df.name, eosUnit)
}

func (df *DataHandlerFactoryPlugin) DeliverUnit(unit common.CmUnit) {
	if unit == nil {
		return
	}

	// Extract buffer from input unit
	cmBuf, isCmBuf := unit.GetBuf().(common.CmBuf)
	if isCmBuf {
		pid, isPidInt := unit.GetField("id").(int)
		if !isPidInt {
			panic("Something wrong with the data")
		}

		_, hasPid := df.handlers[pid]
		field, hasField := cmBuf.GetField("streamType")
		if !hasField {
			return
		}
		dType, ok := field.(int)
		if !ok {
			return
		}
		if !hasPid {
			if dType == 2 {
				df.handlers[pid] = video.MPEG2VideoHandler(pid)
			} else if dType == 129 || dType == 135 {
				df.handlers[pid] = audio.AC3Handler(pid)
			}
		}
		if h, hasHandle := df.handlers[pid]; hasHandle {
			h.Feed(unit)
		}
	}

	df.outputUnit = append(df.outputUnit, unit)

	// Directly output the unit
	reqUnit := common.MakeReqUnit(nil, common.FETCH_REQUEST)
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

func (df *DataHandlerFactoryPlugin) IsRoot() bool {
	return false
}

func (df *DataHandlerFactoryPlugin) Name() string {
	return df.name
}

func DataHandlerFactory(name string) common.IPlugin {
	rv := DataHandlerFactoryPlugin{name: name}
	return &rv
}
