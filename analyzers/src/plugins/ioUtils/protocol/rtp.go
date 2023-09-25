package protocol

import (
	"github.com/tony-507/analyzers/src/common"
	"github.com/tony-507/analyzers/src/plugins/ioUtils/def"
)

type RtpProtocolParser struct {
	logger common.Log
}

func (rtp *RtpProtocolParser) Parse(rawBuf []byte) []def.ParseResult {
	res := make([]def.ParseResult, 1)
	fields := make(map[string]int64)

	r := common.GetBufferReader(rawBuf)
	// RTP header
	def.AssertIntEqual("version", 2, r.ReadBits(2))
	bPad := r.ReadBits(1) != 0
	bExtension := r.ReadBits(1) != 0
	csrcCount := r.ReadBits(4)
	r.ReadBits(1) // Marker
	fields["payloadType"] = int64(r.ReadBits(7))
	fields["seqNumber"] = int64(r.ReadBits(16))
	fields["timestamp"] = int64(r.ReadBits(32))
	fields["syncId"] = int64(r.ReadBits(32))

	for i := 0; i < csrcCount; i++ {
		r.ReadBits(32)
	}

	if bExtension {
		panic("RTP extension header not supported")
	}

	remainedBuf := r.GetRemainedBuffer()
	nPad := 0
	if bPad {
		nPad = int(remainedBuf[len(remainedBuf) - 1])
	}
	res[0] = def.ParseResult{
		Buffer: remainedBuf[:(len(remainedBuf) - nPad)],
		Fields: fields,
	}

	return res
}

func RtpParser() def.IParser {
	return &RtpProtocolParser{
		logger: common.CreateLogger("RTP"),
	}
}
