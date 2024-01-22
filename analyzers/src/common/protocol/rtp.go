package protocol

import (
	"github.com/tony-507/analyzers/src/common/io"
	"github.com/tony-507/analyzers/src/logging"
)

type RtpProtocolParser struct {
	logger logging.Log
}

func (rtp *RtpProtocolParser) Parse(data *ParseResult) []ParseResult {
	rawBuf := data.GetBuffer()
	res := make([]ParseResult, 1)
	fields := make(map[string]int64)

	r := io.GetBufferReader(rawBuf)
	// RTP header
	AssertIntEqual("version", 2, r.ReadBits(2))
	bPad := r.ReadBits(1) != 0
	bExtension := r.ReadBits(1) != 0
	csrcCount := r.ReadBits(4)
	fields["marker"] = int64(r.ReadBits(1))
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
	res[0] = ParseResult{
		Buffer: remainedBuf[:(len(remainedBuf) - nPad)],
		Fields: fields,
	}

	return res
}

func RtpParser() IParser {
	return &RtpProtocolParser{
		logger: logging.CreateLogger("RTP"),
	}
}
