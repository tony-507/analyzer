package st2110

import (
	"strings"
	"time"

	"github.com/tony-507/analyzers/src/plugins/common"
	"github.com/tony-507/analyzers/src/logging"
	"github.com/tony-507/analyzers/src/plugins/common/protocol"
	"github.com/tony-507/analyzers/src/plugins/baseband/def"
	"github.com/tony-507/analyzers/src/tttKernel"
)

type St2110Core struct {
	callback   def.ProcessorCallback
	logger     logging.Log
	inited     bool
	processors map[string]*processor
	rtp        protocol.IParser
	startTime  int64 // Delay some time for probe before start
}

func (core *St2110Core) SetCallback(callback def.ProcessorCallback) {
	core.callback = callback
}

func (core *St2110Core) Feed(unit tttKernel.CmUnit, inputId string) {
	if _, hasKey := core.processors[inputId]; !hasKey {
		core.processors[inputId] = newProcessor(core, inputId)
	}

	pkt := newRtpPacket(unit.GetBuf().GetBuf())

	field, ok := unit.GetBuf().GetField("realtimeInUs")
	var realtime int64

	if !ok {
		realtime = time.Now().UnixMilli()
	} else {
		realtime, _ = field.(int64)
		realtime /= 1000
	}

	core.processors[inputId].tick(time.Duration(realtime) * time.Millisecond)
	core.processors[inputId].feed(&pkt)
}

func (core *St2110Core) PrintInfo(sb *strings.Builder) {
	for _, proc := range core.processors {
		proc.printInfo(sb)
	}
}

func (core *St2110Core) DeliverData(unit *common.MediaUnit) {
	core.callback.OnDataReady(unit)
}

func St2110ProcessorCore() def.ProcessorCore {
	return &St2110Core{
		logger: logging.CreateLogger("St2110Core"),
		inited: false,
		processors: map[string]*processor{},
		rtp: protocol.GetParser(protocol.PROT_RTP),
	}
}
