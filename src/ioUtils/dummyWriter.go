package ioUtils

import "github.com/tony-507/analyzers/src/common"

type DummyWriter struct {
	dummyOut *int
}

func (dw *DummyWriter) setup(writerParam IOWriterParam) {
	dw.dummyOut = writerParam.dummyOut
}

func (dw *DummyWriter) stop() {
	// Do nothing
}

func (dw *DummyWriter) processUnit(unit common.CmUnit) {
	v, _ := unit.GetBuf().(int)
	*dw.dummyOut = (*dw.dummyOut)*10 + v
}
