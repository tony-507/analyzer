package worker

import (
	"github.com/tony-507/analyzers/src/common"
	"github.com/tony-507/analyzers/src/ioUtils"
)

type inputReaderPlugin struct {
	name     string
	impl     ioUtils.FileReader
	callback *Worker
}

func (pg *inputReaderPlugin) SetParameter(m_parameter interface{}) {
	pg.impl.SetParameter(m_parameter)
}

func (pg *inputReaderPlugin) DeliverUnit(unit common.CmUnit) (bool, error) {
	outUnit := pg.impl.DeliverUnit(unit)
	pg.callback.PostRequest(pg.name, outUnit)
	return true, nil
}

func (pg *inputReaderPlugin) FetchUnit() common.CmUnit {
	return pg.impl.FetchUnit()
}

func (pg *inputReaderPlugin) StopPlugin() {
	pg.impl.StopPlugin()
}

func (pg *inputReaderPlugin) SetCallback(callback *Worker) {
	pg.callback = callback
}

func GetInputReaderPlugin(name string) inputReaderPlugin {
	pg := inputReaderPlugin{name: name}
	pg.impl = ioUtils.GetReader(name)
	return pg
}
