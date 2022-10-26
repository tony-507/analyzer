package worker

import (
	"github.com/tony-507/analyzers/src/common"
	"github.com/tony-507/analyzers/src/ioUtils"
	"github.com/tony-507/analyzers/src/resources"
)

type fileWriterPlugin struct {
	name     string
	impl     ioUtils.OutputWriter
	callback *Worker
}

func (pg *fileWriterPlugin) SetParameter(m_parameter interface{}) {
	pg.impl.SetParameter(m_parameter)
}

func (pg *fileWriterPlugin) SetResource(resourceLoader *resources.ResourceLoader) {
	// pg.impl.SetResource(resourceLoader)
}

func (pg *fileWriterPlugin) DeliverUnit(unit common.CmUnit) (bool, error) {
	outUnit := pg.impl.DeliverUnit(unit)
	pg.callback.PostRequest(pg.name, outUnit)
	return true, nil
}

func (pg *fileWriterPlugin) FetchUnit() common.CmUnit {
	return pg.impl.FetchUnit()
}

func (pg *fileWriterPlugin) StartSequence() {
	pg.impl.StartSequence()
}

func (pg *fileWriterPlugin) EndSequence() {
	pg.impl.EndSequence()
	eosUnit := common.MakeReqUnit(pg.name, common.EOS_REQUEST)
	pg.callback.PostRequest(pg.name, eosUnit)
}

func (pg *fileWriterPlugin) SetCallback(callback *Worker) {
	pg.callback = callback
}

func GetFileWriterPlugin(name string) fileWriterPlugin {
	pg := fileWriterPlugin{name: name}
	pg.impl = ioUtils.GetOutputWriter(name)
	return pg
}
