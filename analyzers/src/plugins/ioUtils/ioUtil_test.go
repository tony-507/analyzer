package ioUtils

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tony-507/analyzers/src/logging"
	"github.com/tony-507/analyzers/src/plugins/common/protocol"
	"github.com/tony-507/analyzers/src/plugins/ioUtils/fileReader"
	"github.com/tony-507/analyzers/src/tttKernel"
)

func TestReaderSetParameter(t *testing.T) {
	specs := []string{
		"{\"Uri\":\"file://dummy.ts\"}",
		"{\"Uri\":\"file://hello.abc.ts\"}",
		"{\"Uri\":\"file://hello.abc\"}",
	}

	for _, param := range specs {
		fr := inputReaderPlugin{name: "dummy", logger: logging.CreateLogger("dummy")}
		fr.SetParameter(param)

		_, isFileReader := fr.impl.(*fileReader.FileReaderStruct)
		assert.Equal(t, true, isFileReader, "impl should be a file reader")
		assert.Equal(t, -1, fr.param.maxInCnt, "Uninitialized maxInCnt should be -1")
	}
}

func TestReaderDeliverUnit(t *testing.T) {
	specs := []string{
		"{\"Uri\": \"dummy://\"}",
		"{\"Uri\": \"dummy://\",\"SkipCnt\":2}",  // Deliver with skipping does not change behaviour
		"{\"Uri\": \"dummy://\",\"MaxInCnt\":2}", // Deliver with max input count
	}

	expectedDeliverCnt := []int{5, 5, 2}

	for idx, param := range specs {
		ir := inputReaderPlugin{name: "dummy", logger: logging.CreateLogger("dummy")}
		ir.SetParameter(param)

		ir.SetCallback(func(s string, reqType tttKernel.WORKER_REQUEST, obj interface{}) {
			expected := tttKernel.MakeReqUnit(ir.name, tttKernel.FETCH_REQUEST)
			assert.Equal(t, expected, obj, fmt.Sprintf("[%d] Expect a fetch request", idx))
		})

		for i := 0; i < expectedDeliverCnt[idx]; i++ {
			ir.start()
		}

		ir.SetCallback(func(s string, reqType tttKernel.WORKER_REQUEST, obj interface{}) {
			expected := tttKernel.MakeReqUnit(ir.name, tttKernel.EOS_REQUEST)
			assert.Equal(t, expected, obj, "Expect an EOS request")
		})
		ir.start()
	}
}

func TestTsParser(t *testing.T) {
	parser := protocol.TsParser()
	data := make([]byte, protocol.TS_PKT_SIZE*7)
	for i := 0; i < 7; i++ {
		for j := 0; j < protocol.TS_PKT_SIZE; j++ {
			data[i*protocol.TS_PKT_SIZE+j] = byte(i)
		}
	}
	resList := parser.Parse(&protocol.ParseResult{Buffer: data})
	for idx, res := range resList {
		assert.Equal(t, byte(idx), res.GetBuffer()[0], "Packet value not equal")
	}
}

func TestRtpParser(t *testing.T) {
	data := []byte{
		0x80, 0x60, 0xf2, 0xf6, 0xe4, 0x1a, 0xf0, 0x29, 0xab, 0xcd, 0xab, 0xcd,
		0x01, 0x02, 0x03, 0x04, 0x05,
	}
	resList := protocol.ParseWithParsers([]protocol.IParser{protocol.GetParser(protocol.PROT_RTP)}, &protocol.ParseResult{Buffer: data})
	assert.Equal(t, 1, len(resList))

	res := resList[0]
	timestamp, _ := res.GetField("timestamp")
	assert.Equal(t, int64(3826970665), timestamp, "RTP timestamp not match")
}

func TestParseWithParsers(t *testing.T) {
	// Ensure no infinite loop or weird stuff
	data := make([]byte, protocol.TS_PKT_SIZE*7)
	for i := 0; i < 7; i++ {
		for j := 0; j < protocol.TS_PKT_SIZE; j++ {
			data[i*protocol.TS_PKT_SIZE+j] = byte(i)
		}
	}
	resList := protocol.ParseWithParsers([]protocol.IParser{protocol.GetParser(protocol.PROT_TS), protocol.GetParser(protocol.PROT_TS)}, &protocol.ParseResult{Buffer: data})
	for idx, res := range resList {
		assert.Equal(t, byte(idx), res.GetBuffer()[0], "Packet value not equal")
	}
}
