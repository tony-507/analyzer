package tsdemux

import (
	"github.com/tony-507/analyzers/test/testUtils"
	"github.com/tony-507/analyzers/src/common"
	"errors"
)

func tsDemuxerHandleUnitTest() testUtils.Testcase {
	tc := testUtils.GetTestCase("tsDemuxerHandlerTest", 0)

	tc.Describe("Demuxer should send fetch request once unit available", func (input interface{}) (interface{}, error) {
		m_pMux := GetTsDemuxer("dummy")
		m_parameter := DemuxParams{Mode: DEMUX_DUMMY}
		m_pMux.SetParameter(m_parameter)

		for i := 0; i < 2; i++ {
			dummy := common.IOUnit{Buf: i, IoType: 1, Id: 0}
			rv := m_pMux.DeliverUnit(dummy)
			if i == 0 && rv != nil {
				return nil, errors.New("Non-empty output when demuxer not ready")
			} else if i != 0 {
				expected := common.MakeReqUnit("dummy", common.FETCH_REQUEST)
				return nil, testUtils.Assert_obj_equal(expected, rv)
			}
		}
		return nil, nil
	})

	return tc
}

func AddUnitTestSuite(t *testUtils.Tester) {
	tmg := testUtils.GetTestCaseMgr()

	tmg.AddTest(tsDemuxerHandleUnitTest, []string{"avContainer"})

	t.AddSuite("unitTest", tmg)
}
