package protocol

import (
	"github.com/tony-507/analyzers/src/common/logging"
	"github.com/tony-507/analyzers/src/plugins/ioUtils/def"
)

const (
	TS_PKT_SIZE int = 188
)

type TsProtocolParser struct {
	logger   logging.Log
	count    int
}

func (ts *TsProtocolParser) Parse(data *def.ParseResult) []def.ParseResult {
	rawBuf := data.GetBuffer()
	res := []def.ParseResult{}
	nPackets := len(rawBuf) / TS_PKT_SIZE
	for i := 0; i < nPackets; i++ {
		res = append(res,
			def.ParseResult{
				Buffer: rawBuf[(i * TS_PKT_SIZE):((i + 1) * TS_PKT_SIZE)],
			},
		)
	}
	return res
}

func TsParser() def.IParser {
	return &TsProtocolParser{logger: logging.CreateLogger("TsParser")}
}