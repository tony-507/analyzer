package tsdemux

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tony-507/analyzers/src/common"
	"github.com/tony-507/analyzers/src/plugins/avContainer/model"
)

func TestDemuxDeliverUnit(t *testing.T) {
	m_pMux := TsDemuxer{name: "dummy"}
	m_parameter := "{\"Mode\": \"_DEMUX_DUMMY\"}"
	m_pMux.setParameter(m_parameter)

	m_pMux.setCallback(func(s string, reqType common.WORKER_REQUEST, obj interface{}) {
		expected := common.MakeReqUnit("dummy", common.FETCH_REQUEST)
		assert.Equal(t, expected, obj, "Unit not equal")
	})

	for i := 0; i < 2; i++ {
		dummy := common.MakeIOUnit(i, 1, 0)
		m_pMux.deliverUnit(dummy)
	}
}

func TestDemuxPipeProcessing(t *testing.T) {
	dummyPAT := []byte{0x47, 0x40, 0x00, 0x14, 0x00, 0x00, 0xB0, 0x0D, 0x11, 0x11, 0xC1,
		0x00, 0x00, 0x00, 0x0A, 0xE1, 0x02, 0xAA, 0x4A, 0xE2, 0xD2}

	control := getControl()
	impl := getDemuxPipe(control)

	impl.processUnit(dummyPAT, 0)

	programMap := make(map[int]int, 0)
	programMap[258] = 10
	expectedPAT := model.CreatePAT(0, 4369, 0, true, programMap, 10)

	assert.Equal(t, expectedPAT, impl.content, "PAT not match")

	dummyPMT := []byte{0x47, 0x41, 0x02, 0x14, 0x00, 0x02, 0xb0, 0x1d, 0x00, 0x0a, 0xc1,
		0x00, 0x00, 0xe0, 0x20, 0xf0, 0x00, 0x02, 0xe0, 0x20,
		0xf0, 0x00, 0x04, 0xe0, 0x21, 0xf0, 0x06, 0x0a, 0x04,
		0x65, 0x6e, 0x67, 0x00, 0x75, 0xff, 0x59, 0x3a}

	impl.processUnit(dummyPMT, 0)

	expectedProgDesc := make([]model.Descriptor, 0)

	videoStream := model.DataStream{StreamPid: 32, StreamType: 2, StreamDesc: make([]model.Descriptor, 0)}
	audioStream := model.DataStream{StreamPid: 33, StreamType: 4, StreamDesc: []model.Descriptor{{Tag: 10, Content: "65 6e 67 0"}}}
	streams := []model.DataStream{videoStream, audioStream}

	expectedPmt := model.CreatePMT(258, 2, 10, 0, true, expectedProgDesc, streams, -1)

	assert.Equal(t, expectedPmt, impl.programs[10], "PMT not match")

	unknownPkt := []byte{0x47, 0x40, 0x51, 0x31, 0xff}

	impl.processUnit(unknownPkt, 0)

	assert.Equal(t, 2, impl.control.pCnt, "Process count not match")
}

func tsDemuxPidStatusTest(t *testing.T) {
	dummyPAT := []byte{0x47, 0x40, 0x00, 0x14, 0x00, 0x00, 0xB0, 0x0D, 0x11, 0x11, 0xC1,
		0x00, 0x00, 0x00, 0x0A, 0xE1, 0x02, 0xAA, 0x4A, 0xE2, 0xD2}

	control := getControl()
	impl := getDemuxPipe(control)

	impl.processUnit(dummyPAT, 0)
	assert.Equal(t, 2, len(impl.control.StatusList), "Expect 2 status units: PAT and PMT")
	control.StatusList = make([]common.CmUnit, 0)

	dummyPMT := []byte{0x47, 0x41, 0x02, 0x14, 0x00, 0x02, 0xb0, 0x1d, 0x00, 0x0a, 0xc1,
		0x00, 0x00, 0xe0, 0x20, 0xf0, 0x00, 0x02, 0xe0, 0x20,
		0xf0, 0x00, 0x04, 0xe0, 0x21, 0xf0, 0x06, 0x0a, 0x04,
		0x65, 0x6e, 0x67, 0x00, 0x75, 0xff, 0x59, 0x3a}

	impl.processUnit(dummyPMT, 0)

	assert.Equal(t, 2, len(impl.control.StatusList), "Expect 2 status units: video and audio")
}
