package worker

import (
	"github.com/tony-507/analyzers/src/common"
	"github.com/tony-507/analyzers/src/ioUtils"
)

type fileWriterPlugin struct {
	name     string
	impl     ioUtils.FileWriter
	callback *Worker
}

func (pg *fileWriterPlugin) SetParameter(m_parameter interface{}) {
	pg.impl.SetParameter(m_parameter)
}

func (pg *fileWriterPlugin) DeliverUnit(unit common.CmUnit) (bool, error) {
	outUnit := pg.impl.DeliverUnit(unit)
	pg.callback.PostRequest(pg.name, outUnit)
	return true, nil
}

func (pg *fileWriterPlugin) FetchUnit() common.CmUnit {
	return pg.impl.FetchUnit()
}

func (pg *fileWriterPlugin) StopPlugin() {
	pg.impl.StopPlugin()
}

func (pg *fileWriterPlugin) SetCallback(callback *Worker) {
	pg.callback = callback
}

func GetFileWriterPlugin(name string) fileWriterPlugin {
	pg := fileWriterPlugin{name: name}
	pg.impl = ioUtils.GetFileWriter(name)
	return pg
}
