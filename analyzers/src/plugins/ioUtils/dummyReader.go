package ioUtils

import (
	"github.com/tony-507/analyzers/src/plugins/common/protocol"
	"github.com/tony-507/analyzers/src/plugins/ioUtils/def"
)

type dummyReader struct {
	readCnt int
}

func (dr *dummyReader) Setup(config def.IReaderConfig) {
	dr.readCnt = 0
}

func (dr *dummyReader) StartRecv() error {
	return nil
}

func (dr *dummyReader) StopRecv() error {
	return nil
}

func (dr *dummyReader) DataAvailable() (protocol.ParseResult, bool) {
	dr.readCnt += 1
	if dr.readCnt > 5 {
		return protocol.ParseResult{}, false
	}
	return protocol.EmptyResult(), true
}
