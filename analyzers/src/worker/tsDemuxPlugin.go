package worker

import (
	"github.com/tony-507/analyzers/src/avContainer/tsdemux"
	"github.com/tony-507/analyzers/src/common"
	"github.com/tony-507/analyzers/src/resources"
)

type tsDemuxPlugin struct {
	name     string
	impl     tsdemux.TsDemuxer
	callback *Worker
}

func (pg *tsDemuxPlugin) SetParameter(m_parameter interface{}) {
	pg.impl.SetParameter(m_parameter)
}

func (pg *tsDemuxPlugin) SetResource(resourceLoader *resources.ResourceLoader) {
	pg.impl.SetResource(resourceLoader)
}

func (pg *tsDemuxPlugin) DeliverUnit(unit common.CmUnit) (bool, error) {
	outUnit := pg.impl.DeliverUnit(unit)
	pg.callback.PostRequest(pg.name, outUnit)
	return true, nil
}

func (pg *tsDemuxPlugin) FetchUnit() common.CmUnit {
	return pg.impl.FetchUnit()
}

func (pg *tsDemuxPlugin) StartSequence() {
	pg.impl.StartSequence()
}

func (pg *tsDemuxPlugin) EndSequence() {
	pg.impl.EndSequence()
	eosUnit := common.MakeReqUnit(pg.name, common.EOS_REQUEST)
	pg.callback.PostRequest(pg.name, eosUnit)
}

func (pg *tsDemuxPlugin) SetCallback(callback *Worker) {
	pg.callback = callback
}

func GetTsDemuxPlugin(name string) tsDemuxPlugin {
	pg := tsDemuxPlugin{name: name}
	pg.impl = tsdemux.GetTsDemuxer(name)
	return pg
}
