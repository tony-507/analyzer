package worker

import (
	"github.com/tony-507/analyzers/src/avContainer"
	"github.com/tony-507/analyzers/src/common"
)

type tsDemuxPlugin struct {
	name     string
	impl     avContainer.TsDemuxer
	callback *Worker
}

func (pg *tsDemuxPlugin) SetParameter(m_parameter interface{}) {
	pg.impl.SetParameter(m_parameter)
}

func (pg *tsDemuxPlugin) DeliverUnit(unit common.CmUnit) (bool, error) {
	outUnit := pg.impl.DeliverUnit(unit)
	pg.callback.PostRequest(pg.name, outUnit)
	return true, nil
}

func (pg *tsDemuxPlugin) FetchUnit() common.CmUnit {
	return pg.impl.FetchUnit()
}

func (pg *tsDemuxPlugin) StopPlugin() {
	pg.impl.StopPlugin()
}

func (pg *tsDemuxPlugin) SetCallback(callback *Worker) {
	pg.callback = callback
}

func GetTsDemuxPlugin(name string) tsDemuxPlugin {
	pg := tsDemuxPlugin{name: name}
	pg.impl = avContainer.GetTsDemuxer(name)
	return pg
}
