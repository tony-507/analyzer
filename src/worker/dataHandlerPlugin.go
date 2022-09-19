package worker

import (
	"github.com/tony-507/analyzers/src/common"
	datahandler "github.com/tony-507/analyzers/src/dataHandler"
)

type DataHandlerPlugin struct {
	name     string
	impl     datahandler.DataHandlerFactory
	callback *Worker
}

func (pg *DataHandlerPlugin) SetParameter(m_parameter interface{}) {
	pg.impl.SetParameter(m_parameter)
}

func (pg *DataHandlerPlugin) DeliverUnit(unit common.CmUnit) (bool, error) {
	outUnit := pg.impl.DeliverUnit(unit)
	pg.callback.PostRequest(pg.name, outUnit)
	return true, nil
}

func (pg *DataHandlerPlugin) FetchUnit() common.CmUnit {
	return pg.impl.FetchUnit()
}

func (pg *DataHandlerPlugin) StopPlugin() {
	pg.impl.StopPlugin()
}

func (pg *DataHandlerPlugin) SetCallback(callback *Worker) {
	pg.callback = callback
}

func GetDataHandlerPlugin(name string) DataHandlerPlugin {
	pg := DataHandlerPlugin{name: name}
	pg.impl = datahandler.GetDataHandlerFactory(name)
	return pg
}
