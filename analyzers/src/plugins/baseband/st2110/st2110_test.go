package st2110

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/tony-507/analyzers/src/common"
	"github.com/tony-507/analyzers/src/plugins/baseband/def"
	"github.com/tony-507/analyzers/src/tttKernel"
)

type dummyCallback struct {}

func (cb *dummyCallback) Feed(tttKernel.CmUnit, string) {}

func (cb *dummyCallback) PrintInfo(*strings.Builder) {}

func (cb *dummyCallback) SetCallback(def.ProcessorCallback) {}

func (cb *dummyCallback) DeliverData(*common.MediaUnit) {}

func TestStreamProbe(t *testing.T) {
	testcases := map[string]map[string]int{
		"Video": {
			"cnt": 1000,
			"sampleRate": 90000,
			"expected": int(_ST211020),
		},
		"Audio": {
			"cnt": 1000,
			"sampleRate": 44100,
			"expected": int(_ST211030),
		},
		"Data": {
			"cnt": 100,
			"sampleRate": 90000,
			"expected": int(_ST211040),
		},
	}
	for caseName, testcase := range testcases {
		t.Run(caseName, func(t *testing.T) {
			cb := &dummyCallback{}
			proc := newProcessor(cb, "test")

			initRtp := 10
			initUtc := 100 * time.Millisecond
			pktCnt := testcase["cnt"]
			sampleRate := testcase["sampleRate"]

			for i := 0; i < pktCnt; i++ {
				rtp := initRtp + i * sampleRate / pktCnt
				utc := initUtc + time.Duration(i) * time.Second / time.Duration(pktCnt)
				proc.tick(utc)

				pkt := rtpPacket{
					payloadType: 1,
					timestamp: uint32(rtp),
					marker: false,
				}
				proc.feed(&pkt)
			}

			assert.Equal(t, testcase["expected"], int(proc.streamType))
		})
	}
}
