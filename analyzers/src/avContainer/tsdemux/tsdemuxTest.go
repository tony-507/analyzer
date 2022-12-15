package tsdemux

import (
	"github.com/tony-507/analyzers/src/avContainer/model"
	"github.com/tony-507/analyzers/src/common"
	"github.com/tony-507/analyzers/test/testUtils"
)

func tsDemuxerDeliverUnitTest() testUtils.Testcase {
	tc := testUtils.GetTestCase("tsDemuxerDeliverUnitTest", 0)

	tc.Describe("Demuxer should send fetch request once unit available", func(input interface{}) (interface{}, error) {
		m_pMux := GetTsDemuxer("dummy")
		m_parameter := DemuxParams{Mode: DEMUX_DUMMY}
		m_pMux.SetParameter(m_parameter)

		m_pMux.SetCallback(func(s string, reqType common.WORKER_REQUEST, obj interface{}) {
			expected := common.MakeReqUnit("dummy", common.FETCH_REQUEST)
			err := testUtils.Assert_obj_equal(expected, obj)
			if err != nil {
				panic(err)
			}
		})

		for i := 0; i < 2; i++ {
			dummy := common.IOUnit{Buf: i, IoType: 1, Id: 0}
			m_pMux.DeliverUnit(dummy)
		}
		return nil, nil
	})

	return tc
}

func tsDemuxPipeProcessTest() testUtils.Testcase {
	tc := testUtils.GetTestCase("tsDemuxPipeProcessTest", 0)

	tc.Describe("DemuxPipe is able to set PAT", func(input interface{}) (interface{}, error) {
		dummyPAT := []byte{0x47, 0x40, 0x00, 0x14, 0x00, 0x00, 0xB0, 0x0D, 0x11, 0x11, 0xC1,
			0x00, 0x00, 0x00, 0x0A, 0xE1, 0x02, 0xAA, 0x4A, 0xE2, 0xD2}

		control := getControl()
		impl := getDemuxPipe(control)

		impl.processUnit(dummyPAT, 0)

		programMap := make(map[int]int, 0)
		programMap[258] = 10
		expectedPAT := model.CreatePAT(0, 4369, 0, true, programMap, 10)

		err := testUtils.Assert_obj_equal(expectedPAT, impl.content)

		return impl, err
	})

	tc.Describe("DemuxPipe is able to set PMT", func(input interface{}) (interface{}, error) {
		dummyPMT := []byte{0x47, 0x41, 0x02, 0x14, 0x00, 0x02, 0xb0, 0x1d, 0x00, 0x0a, 0xc1,
			0x00, 0x00, 0xe0, 0x20, 0xf0, 0x00, 0x02, 0xe0, 0x20,
			0xf0, 0x00, 0x04, 0xe0, 0x21, 0xf0, 0x06, 0x0a, 0x04,
			0x65, 0x6e, 0x67, 0x00, 0x75, 0xff, 0x59, 0x3a}

		impl, _ := input.(tsDemuxPipe)
		impl.processUnit(dummyPMT, 0)

		expectedProgDesc := make([]model.Descriptor, 0)

		videoStream := model.DataStream{StreamPid: 32, StreamType: 2, StreamDesc: make([]model.Descriptor, 0)}
		audioStream := model.DataStream{StreamPid: 33, StreamType: 4, StreamDesc: []model.Descriptor{{Tag: 10, Content: "65 6e 67 0"}}}
		streams := []model.DataStream{videoStream, audioStream}

		expectedPmt := model.CreatePMT(258, 2, 10, 0, true, expectedProgDesc, streams, -1)

		err := testUtils.Assert_obj_equal(expectedPmt, impl.programs[0])
		return impl, err
	})

	tc.Describe("DemuxPipe passes through unknown packet", func(input interface{}) (interface{}, error) {
		unknownPkt := []byte{0x47, 0x40, 0x51, 0x31, 0xff}

		impl, _ := input.(tsDemuxPipe)
		impl.processUnit(unknownPkt, 0)

		err := testUtils.Assert_equal(2, impl.control.pCnt)
		return impl, err
	})

	return tc
}

func AddUnitTestSuite(t *testUtils.Tester) {
	tmg := testUtils.GetTestCaseMgr()

	tmg.AddTest(tsDemuxerDeliverUnitTest, []string{"avContainer"})
	tmg.AddTest(tsDemuxPipeProcessTest, []string{"avContainer"})

	t.AddSuite("unitTest", tmg)
}
