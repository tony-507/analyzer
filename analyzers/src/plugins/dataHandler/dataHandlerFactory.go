package dataHandler

import (
	"github.com/tony-507/analyzers/src/common"
	"github.com/tony-507/analyzers/src/logs"
	"github.com/tony-507/analyzers/src/plugins/dataHandler/utils"
	"github.com/tony-507/analyzers/src/plugins/dataHandler/video"
)

type DataHandlerFactory struct {
	logger     logs.Log
	callback   common.RequestHandler
	handlers   map[int]utils.DataHandler
	outputUnit []common.CmUnit
	isRunning  bool
	name       string
}

func (df *DataHandlerFactory) setCallback(callback common.RequestHandler) {
	df.callback = callback
}

func (df *DataHandlerFactory) setParameter(m_parameter string) {
	df._setup()
}

func (df *DataHandlerFactory) setResource(loader *common.ResourceLoader) {}

func (df *DataHandlerFactory) _setup() {
	df.logger = logs.CreateLogger("DataHandlerFactory")
	df.handlers = map[int]utils.DataHandler{}
	df.outputUnit = []common.CmUnit{}
	df.isRunning = true
}

func (df *DataHandlerFactory) startSequence() {
	df.logger.Log(logs.INFO, "Data handler factory is started")
}

func (df *DataHandlerFactory) endSequence() {
	df.logger.Log(logs.INFO, "Data handler factory is stopped")
	eosUnit := common.MakeReqUnit(df.name, common.EOS_REQUEST)
	common.Post_request(df.callback, df.name, eosUnit)
}

func (df *DataHandlerFactory) deliverUnit(unit common.CmUnit) {
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
		dType, hasField := cmBuf.GetField("dataType")
		if !hasField {
			return
		}
		if !hasPid {
			df.logger.Log(logs.INFO, "Receive pid %d at dataHandlerFactory", pid)
			if dType == "MPEG 2 video" {
				df.handlers[pid] = video.MPEG2VideoHandler()
			}
		}
		if dType == "MPEG 2 video" {
			df.handlers[pid].Feed(unit)
		}
	}

	df.outputUnit = append(df.outputUnit, unit)

	// Directly output the unit
	reqUnit := common.MakeReqUnit(nil, common.FETCH_REQUEST)
	common.Post_request(df.callback, df.name, reqUnit)
}

func (df *DataHandlerFactory) deliverStatus(unit common.CmUnit) {}

func (df *DataHandlerFactory) fetchUnit() common.CmUnit {
	if len(df.outputUnit) == 0 {
		return nil
	} else if len(df.outputUnit) == 1 {
		rv := df.outputUnit[0]
		df.outputUnit = make([]common.CmUnit, 0)
		return rv
	} else {
		rv := df.outputUnit[0]
		df.outputUnit = df.outputUnit[1:]
		return rv
	}
}

func GetDataHandlerFactory(name string) common.Plugin {
	rv := DataHandlerFactory{name: name}
	return common.CreatePlugin(
		name,
		false,
		rv.setCallback,
		rv.setParameter,
		rv.setResource,
		rv.startSequence,
		rv.deliverUnit,
		rv.deliverStatus,
		rv.fetchUnit,
		rv.endSequence,
	)
}
