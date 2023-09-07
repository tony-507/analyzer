package ioUtils

import (
	"github.com/tony-507/analyzers/src/common"
)

type dummyReader struct {
	readCnt int
}

func (dr *dummyReader) Setup() {
	dr.readCnt = 0
}

func (dr *dummyReader) StartRecv() error {
	return nil
}

func (dr *dummyReader) StopRecv() error {
	return nil
}

func (dr *dummyReader) DataAvailable(unit *common.IOUnit) bool {
	dr.readCnt += 1
	if dr.readCnt > 5 {
		return false
	}
	unit.IoType = 3
	unit.Buf = dr.readCnt
	return true
}
