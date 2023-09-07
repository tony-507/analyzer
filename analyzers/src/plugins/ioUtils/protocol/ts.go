package protocol

import (
	"github.com/tony-507/analyzers/src/common"
	"github.com/tony-507/analyzers/src/plugins/ioUtils/def"
)

const (
	TS_PKT_SIZE int = 188
)

type TsPacket struct {
	buffer []byte
}

func (pkt TsPacket) GetBuffer() []byte {
	return pkt.buffer
}

func (pkt TsPacket) GetField(field string) (int, bool) {
	return 0, false
}

type TsProtocolParser struct {
	logger   common.Log
	count    int
}

func (ts *TsProtocolParser) Parse(rawBuf []byte) []def.ParseResult {
	res := []def.ParseResult{}
	nPackets := len(rawBuf) / TS_PKT_SIZE
	for i := 0; i < nPackets; i++ {
		res = append(res,
			TsPacket{buffer: rawBuf[(i * TS_PKT_SIZE):((i + 1) * TS_PKT_SIZE)]},
		)
	}
	return res
}

func TsParser() def.IParser {
	return &TsProtocolParser{logger: common.CreateLogger("TsParser")}
}