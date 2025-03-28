package ioUtils

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/tony-507/analyzers/src/logging"
	"github.com/tony-507/analyzers/src/plugins/common"
	"github.com/tony-507/analyzers/src/plugins/common/clock"
	"github.com/tony-507/analyzers/src/plugins/common/io"
	"github.com/tony-507/analyzers/src/plugins/common/protocol"
	"github.com/tony-507/analyzers/src/plugins/ioUtils/def"
	"github.com/tony-507/analyzers/src/plugins/ioUtils/fileReader"
	"github.com/tony-507/analyzers/src/tttKernel"
)

type inputStat struct {
	outCnt        int
	prevTimestamp int64
	prevTimecode  common.TimeCode
	errCount      int
}

type inputParam struct {
	skipCnt      int
	maxInCnt     int
	dumpRawInput bool
}

type inputReaderPlugin struct {
	name         string
	callback     tttKernel.RequestHandler
	impl         def.IReader
	isRunning    bool
	loader       *tttKernel.ResourceLoader
	logger       logging.Log
	outputQueue  []tttKernel.CmUnit
	stat         inputStat
	param        inputParam
	parsers      []protocol.IParser
	rawBufWriter io.FileWriter
}

func (ir *inputReaderPlugin) StartSequence() {
	ir.isRunning = true

	if ir.param.dumpRawInput {
		ir.rawBufWriter = io.RawWriter(ir.loader.Query("outDir", nil), "rawBuffer")
		if err := ir.rawBufWriter.Open(); err != nil {
			ir.logger.Error("Fail to open file for dumping raw input: %s", err.Error())
		}
	}

	err := ir.impl.StartRecv()
	if err != nil {
		panic(err)
	}
	go ir.DeliverUnit(nil, "")
}

func (ir *inputReaderPlugin) EndSequence() {
	ir.logger.Info("Ending sequence, fetch count = %d", ir.stat.outCnt)
	ir.isRunning = false
	err := ir.impl.StopRecv()
	if err != nil {
		panic(err)
	}

	if ir.rawBufWriter != nil {
		ir.rawBufWriter.Close()
	}
	eosUnit := tttKernel.MakeReqUnit(ir.name, tttKernel.EOS_REQUEST)
	tttKernel.Post_request(ir.callback, ir.name, eosUnit)
}

func (ir *inputReaderPlugin) SetCallback(callback tttKernel.RequestHandler) {
	ir.callback = callback
}

func (ir *inputReaderPlugin) SetParameter(m_parameter string) {
	var param ioReaderParam
	if err := json.Unmarshal([]byte(m_parameter), &param); err != nil {
		panic(err)
	}
	if param.SkipCnt > 0 {
		ir.param.skipCnt = param.SkipCnt
	} else {
		ir.param.skipCnt = 0
	}
	if param.MaxInCnt > 0 {
		ir.param.maxInCnt = param.MaxInCnt
	} else {
		ir.param.maxInCnt = -1
	}

	ir.param.dumpRawInput = param.DumpRawInput

	ir.stat.outCnt = 0
	srcType := "unknown"

	u, e := url.Parse(param.Uri)
	if e != nil {
		panic(e)
	}

	switch u.Scheme {
	case "file":
		srcType = "file"
		ir.impl = fileReader.FileReader(ir.name, u.Path)
	case "udp":
		srcType = "UDP"
		udp := udpInputParam{
			Address: u.Host,
			Itf:     u.Query()["interface"][0],
			Timeout: 0,
		}
		ir.impl = udpReader(&udp, ir.name)
	default:
		srcType = "dummy"
		ir.impl = &dummyReader{}
	}

	if (param.Protocols != "") {
		for _, prot := range strings.Split(param.Protocols, ",") {
			ir.parsers = append(ir.parsers, protocol.GetParser(protocol.StringToProtocol(prot)))
		}
	}

	ir.logger.Info("%s reader created", srcType)

	ir.impl.Setup(def.IReaderConfig{
		Parsers: ir.parsers,
	})
}

func (ir *inputReaderPlugin) SetResource(loader *tttKernel.ResourceLoader) {
	ir.loader = loader
}

func (ir *inputReaderPlugin) DeliverUnit(unit tttKernel.CmUnit, inputId string) {
	if ir.isRunning {
		ir.start()
		reqUnit := tttKernel.MakeReqUnit(ir.name, tttKernel.DELIVER_REQUEST)
		tttKernel.Post_request(ir.callback, ir.name, reqUnit)
	}
}

func (ir *inputReaderPlugin) DeliverStatus(unit tttKernel.CmUnit) {}

func (ir *inputReaderPlugin) start() {
	// Here, we will keep delivering until EOS is signaled
	res, ok := ir.impl.DataAvailable()
	if ir.param.maxInCnt != 0 && ok {
		if !res.IsEmpty {
			ir.stat.outCnt += 1
			ir.param.maxInCnt -= 1

			cmBuf := tttKernel.MakeSimpleBuf(res.GetBuffer())
			newUnit := common.NewMediaUnit(cmBuf, common.UNKNOWN_UNIT)

			ir.processMetadata(cmBuf, &res)

			ir.outputQueue = append(ir.outputQueue, newUnit)
			reqUnit := tttKernel.MakeReqUnit(ir.name, tttKernel.FETCH_REQUEST)
			tttKernel.Post_request(ir.callback, ir.name, reqUnit)
		}
	} else {
		// Stop reader
		eosUnit := tttKernel.MakeReqUnit(ir.name, tttKernel.EOS_REQUEST)
		tttKernel.Post_request(ir.callback, ir.name, eosUnit)
	}
}

func (ir *inputReaderPlugin) processMetadata(cmBuf tttKernel.CmBuf, res *protocol.ParseResult) {
	if realtime, ok := res.GetField("realtimeInUs"); ok {
		cmBuf.SetField("realtimeInUs", realtime, false)
	}

	if timestamp, ok := res.GetField("timestamp"); ok {
		if ir.stat.prevTimestamp != timestamp {
			nextTc := common.GetNextTimeCode(&ir.stat.prevTimecode, 30000, 1001, true)
			tc, err := common.RtpTimestampToTimeCode(clock.MpegClk(timestamp) * clock.Clk90k, -1, 30000, 1001, false, 0)

			// HACK: Cannot identify field and frame right now, so we skip the case for same timecode
			if ir.stat.prevTimecode != tc && nextTc != tc {
				ir.logger.Error("VITC jump detected: %s -> %s. Expected: %s", ir.stat.prevTimecode.ToString(), tc.ToString(), nextTc.ToString())
			}
			ir.logger.Info("%d -> %s", timestamp, tc.ToString())
			ir.stat.prevTimestamp = timestamp
			ir.stat.prevTimecode = tc
			if err != nil {
				ir.stat.errCount++
			}
		}
	}
}

func (ir *inputReaderPlugin) FetchUnit() tttKernel.CmUnit {
	var rv tttKernel.CmUnit

	if len(ir.outputQueue) != 0 && ir.param.skipCnt <= 0 {
		rv = ir.outputQueue[0]
	}

	if len(ir.outputQueue) == 1 {
		ir.outputQueue = make([]tttKernel.CmUnit, 0)
	} else {
		ir.outputQueue = ir.outputQueue[1:]
	}

	ir.param.skipCnt -= 1

	if ir.rawBufWriter != nil && rv != nil {
		ir.rawBufWriter.Write(rv.GetBuf())
	}

	return rv
}

func (ir *inputReaderPlugin) Name() string {
	return ir.name
}

func (ir *inputReaderPlugin) PrintInfo(sb *strings.Builder) {
	stat := ir.stat

	sb.WriteString(fmt.Sprintf("\tOut count: %d\n", stat.outCnt))
	if stat.errCount != 0 {
		sb.WriteString(fmt.Sprintf("\tErr timestamp: %d\n", stat.prevTimestamp))
		sb.WriteString(fmt.Sprintf("\tErr timecode: %s\n", stat.prevTimecode.ToString()))
		sb.WriteString(fmt.Sprintf("\tErr count: %d", stat.errCount))
		stat.errCount = 0
	}
}

func InputReader(name string) tttKernel.IPlugin {
	rv := inputReaderPlugin{
		name: name,
		logger: logging.CreateLogger(name),
		stat: inputStat{
			outCnt: 0,
			prevTimestamp: -1,
			errCount: 0,
		},
		rawBufWriter: nil,
	}
	return &rv
}
