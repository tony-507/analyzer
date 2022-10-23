package ioUtils

import (
	"github.com/tony-507/analyzers/src/common"
)

type dummyReader struct {
	readCnt int
}

func (dr *dummyReader) setup() {
	dr.readCnt = 0
}

func (dr *dummyReader) startRecv() {}

func (dr *dummyReader) stopRecv() {}

func (dr *dummyReader) dataAvailable(unit *common.IOUnit) bool {
	dr.readCnt += 1
	if dr.readCnt > 10 {
		return false
	}
	unit.IoType = 3
	unit.Buf = dr.readCnt
	return true
}
