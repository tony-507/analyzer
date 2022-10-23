package worker

import (
	"time"

	"github.com/tony-507/analyzers/src/common"
	"github.com/tony-507/analyzers/src/ioUtils"
	"github.com/tony-507/analyzers/src/resources"
)

type inputReaderPlugin struct {
	name     string
	impl     ioUtils.InputReader
	callback *Worker
}

func (pg *inputReaderPlugin) SetParameter(m_parameter interface{}) {
	pg.impl.SetParameter(m_parameter)
}

func (pg *inputReaderPlugin) SetResource(resourceLoader *resources.ResourceLoader) {
	// pg.impl.SetResource(resourceLoader)
}

func (pg *inputReaderPlugin) DeliverUnit(unit common.CmUnit) (bool, error) {
	// Wait for few seconds for parameters to be set for all plugins
	time.Sleep(5 * time.Second)

	// Here, we will keep delivering until EOS is signaled by impl
	for {
		outUnit := pg.impl.DeliverUnit(unit)
		pg.callback.PostRequest(pg.name, outUnit)

		reqType, _ := outUnit.GetField("reqType").(common.WORKER_REQUEST)
		if reqType == common.EOS_REQUEST {
			break
		}
	}
	return true, nil
}

func (pg *inputReaderPlugin) FetchUnit() common.CmUnit {
	return pg.impl.FetchUnit()
}

func (pg *inputReaderPlugin) StartSequence() {
	pg.impl.StartSequence()
	// As the first plugin, we need to start receive input after initialization	
	go pg.DeliverUnit(nil)
}

func (pg *inputReaderPlugin) EndSequence() {
	pg.impl.EndSequence()
}

func (pg *inputReaderPlugin) SetCallback(callback *Worker) {
	pg.callback = callback
}

func GetInputReaderPlugin(name string) inputReaderPlugin {
	pg := inputReaderPlugin{name: name}
	pg.impl = ioUtils.GetReader(name)
	return pg
}
