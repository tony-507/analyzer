package avContainer

import (
	"errors"

	"github.com/tony-507/analyzers/test/testUtils"
)

func readPATTest() testUtils.Testcase {
	tc := testUtils.GetTestCase("readPatTest")

	tc.Describe("Initialization", func(input interface{}) (interface{}, error) {
		dummyPAT := []byte{0x00, 0x00, 0xB0, 0x0D, 0x11, 0x11, 0xC1,
			0x00, 0x00, 0x00, 0x0A, 0xE1, 0x02, 0xAA, 0x4A, 0xE2, 0xD2}
		return dummyPAT, nil
	})

	tc.Describe("Parse PAT", func(input interface{}) (interface{}, error) {
		dummyPAT, isBts := input.([]byte)
		if !isBts {
			err := errors.New("Input not passed to next step")
			return nil, err
		}
		PAT, parsingErr := ParsePAT(dummyPAT, 0)
		if parsingErr != nil {
			return nil, parsingErr
		}

		var err error
		err = testUtils.Assert_equal(PAT.tableId, 0)
		if err != nil {
			return nil, err
		}

		err = testUtils.Assert_equal(PAT.tableIdExt, 4369)
		if err != nil {
			return nil, err
		}

		err = testUtils.Assert_equal(PAT.Version, 0)
		if err != nil {
			return nil, err
		}

		return nil, nil
	})

	return tc
}

func readPMTTest() testUtils.Testcase {
	tc := testUtils.GetTestCase("readPmtTest")

	tc.Describe("Initialization", func(input interface{}) (interface{}, error) {
		dummyPMT := []byte{0x00, 0x02, 0xb0, 0x1d, 0x00, 0x0a, 0xc1,
			0x00, 0x00, 0xe0, 0x20, 0xf0, 0x00, 0x02, 0xe0, 0x20,
			0xf0, 0x00, 0x04, 0xe0, 0x21, 0xf0, 0x06, 0x0a, 0x04,
			0x65, 0x6e, 0x67, 0x00, 0x75, 0xff, 0x59, 0x3a}
		return dummyPMT, nil
	})

	tc.Describe("Parse PMT", func(input interface{}) (interface{}, error) {
		dummyPMT, isBts := input.([]byte)
		if !isBts {
			err := errors.New("Input not passed to next step")
			return nil, err
		}

		// Create expected PMT struct
		expectedProgDesc := make([]Descriptor, 0)

		videoStream := DataStream{StreamPid: 32, StreamType: 2, StreamDesc: make([]Descriptor, 0)}
		audioStream := DataStream{StreamPid: 33, StreamType: 4, StreamDesc: []Descriptor{{Tag: 10, Content: "65 6e 67 0"}}}
		streams := []DataStream{videoStream, audioStream}

		expectedPmt := PMT{PktCnt: 1, PmtPid: 258, tableId: 2, ProgNum: 10, Version: 0, curNextIdr: true, ProgDesc: expectedProgDesc, Streams: streams, crc32: -1}

		// Parse
		parsed := ParsePMT(dummyPMT, 258, 1)

		err := testUtils.Assert_obj_equal(expectedPmt, parsed)
		if err != nil {
			return nil, err
		} else {
			return nil, nil
		}
	})

	return tc
}

func AddUnitTestSuite(t *testUtils.Tester) {
	tmg := testUtils.GetTestCaseMgr()

	// We may add custom test filter here later

	// common
	tmg.AddTest(readPATTest, []string{"avContainer"})
	tmg.AddTest(readPMTTest, []string{"avContainer"})

	t.AddSuite("unitTest", tmg)
}
