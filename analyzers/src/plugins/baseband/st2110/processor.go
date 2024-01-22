package st2110

import (
	"fmt"
	"strings"
	"time"

	"github.com/tony-507/analyzers/src/plugins/common"
	"github.com/tony-507/analyzers/src/logging"
	"github.com/tony-507/analyzers/src/plugins/baseband/def"
)

type _TYPE int

const (
	_UNKNOWN  _TYPE = 0
	_ST211020 _TYPE = 1
	_ST211030 _TYPE = 2
	_ST211040 _TYPE = 3
)

type probeInfo struct {
	initRtp    uint32
	initUtc    time.Duration
	curRtp     uint32
	curUtc     time.Duration
	pktCnt     int
	maxRtpDiff int
}

type videoInfo struct {
	isSegmented  bool    // A frame is represented as two fields or segments
	field        bool    // Field flag is used
	fr_num       int
	fr_den       int
}

/*
 * A processor has two stages: Probe and work
 *
 * The processor tries to probe the stream type.
 *
 * The processor tries to probe stream properties.
 *
 * If probe is successful, the processor starts processing the data
 */
type processor struct {
	logger      logging.Log
	callback    def.ProcessorCore
	id          string
	buffer      []rtpPacket
	delay       int // Number of frames to delay
	info        probeInfo
	vInfo       videoInfo
	streamType  _TYPE
}

func (p *processor) feed(inPkt *rtpPacket) {
	if p.info.initRtp == 0 {
		p.info.initRtp = inPkt.timestamp
	}
	p.info.pktCnt++

	curRtp := inPkt.timestamp

	// Empty packet from unit test
	if len(inPkt.payload) != 0 {
		p.buffer = append(p.buffer, *inPkt)
	}

	if curRtp != p.info.curRtp {
		p.info.maxRtpDiff = int(p.getRtpDiff(curRtp, p.info.curRtp))
		p.info.curRtp = curRtp

		p.probeStream()
		p.process()
	}
}

func (p *processor) probeStream() {
	elapsedUtc := p.info.curUtc - p.info.initUtc
	elapsedRtp := p.getRtpDiff(p.info.curRtp, p.info.initRtp)
	if elapsedUtc.Milliseconds() == 0 {
		return
	}

	sampleRate := common.MatchValueInList(
		int(elapsedRtp) * 1000 / int(elapsedUtc.Milliseconds()),
		[]int{48000, 44100, 96000, 90000},
		500,
	)

	streamType := _UNKNOWN
	if sampleRate == 90000 {
		pktRate := int64(p.info.pktCnt) * 1000 / elapsedUtc.Milliseconds()
		if pktRate > 200 {
			streamType = _ST211020
		} else {
			streamType = _ST211040
		}
	} else if sampleRate != -1 {
		streamType = _ST211030
	}

	if p.streamType != streamType {
		if p.streamType != _UNKNOWN {
			p.logger.Error("Stream type changes from %v to %v", p.streamType, streamType)
		} else {
			p.logger.Info("Stream type detected to be %v", streamType)
		}
	}

	p.streamType = streamType
}

// Process at most 1 packet at a time
func (p *processor) process() {
	if len(p.buffer) == 0 {
		return
	}

	if p.streamType == _UNKNOWN {
		return
	}

	switch p.streamType {
	case _ST211020:
		p.processVideoPackets()
	default:
	}
}

func (p *processor) tick(curTime time.Duration) {
	if p.info.initUtc == 0 {
		p.info.initUtc = curTime
	}
	p.info.curUtc = curTime
}

func (p *processor) getRtpDiff(minuend uint32, subtrachend uint32) uint32 {
	// TODO: Handle RTP loop point
	return minuend - subtrachend
}

func (p *processor) printInfo(sb *strings.Builder) {
	sb.WriteString(fmt.Sprintf("%s:\n", p.id))
	sb.WriteString(fmt.Sprintf("\tPacket count: %d\n", p.info.pktCnt))
	sb.WriteString(fmt.Sprintf("\tMax RTP difference: %d\n", p.info.maxRtpDiff))
	sb.WriteString(fmt.Sprintf("\tElapsed RTP timestamp: %d\n", p.info.curRtp - p.info.initRtp))
	sb.WriteString(fmt.Sprintf("\tElapsed UTC: %dms\n", (p.info.curUtc - p.info.initUtc).Milliseconds()))
}

func newProcessor(callback def.ProcessorCore, originator string) *processor {
	return &processor{
		callback: callback,
		logger: logging.CreateLogger(fmt.Sprintf("ST2110_%s", originator)),
		id: originator,
		buffer: []rtpPacket{},
		delay: 3,
		info: probeInfo{
			initRtp: 0,
			initUtc: 0,
			curRtp: 0,
			curUtc: 0,
			pktCnt: 0,
			maxRtpDiff: 0,
		},
		streamType: _UNKNOWN,
	}
}
