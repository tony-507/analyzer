package avContainer

import (
	"github.com/tony-507/analyzers/src/avContainer/model"
	"github.com/tony-507/analyzers/src/avContainer/tsdemux"
	"github.com/tony-507/analyzers/src/testUtils"
)

func readPATTest() testUtils.Testcase {
	tc := testUtils.GetTestCase("readPatTest", 0)

	tc.Describe("Basic PAT parsing", func(input interface{}) (interface{}, error) {
		dummyPAT := []byte{0x00, 0x00, 0xB0, 0x0D, 0x11, 0x11, 0xC1,
			0x00, 0x00, 0x00, 0x0A, 0xE1, 0x02, 0xAA, 0x4A, 0xE2, 0xD2}

		PAT, parsingErr := model.ParsePAT(dummyPAT, 0)
		if parsingErr != nil {
			return nil, parsingErr
		}

		programMap := make(map[int]int, 0)
		programMap[258] = 10
		expected := model.CreatePAT(0, 4369, 0, true, programMap, 10)

		err := testUtils.Assert_obj_equal(expected, PAT)

		return nil, err
	})

	return tc
}

func readPMTTest() testUtils.Testcase {
	tc := testUtils.GetTestCase("readPmtTest", 0)

	tc.Describe("Basic PMT parsing", func(input interface{}) (interface{}, error) {
		dummyPMT := []byte{0x00, 0x02, 0xb0, 0x1d, 0x00, 0x0a, 0xc1,
			0x00, 0x00, 0xe0, 0x20, 0xf0, 0x00, 0x02, 0xe0, 0x20,
			0xf0, 0x00, 0x04, 0xe0, 0x21, 0xf0, 0x06, 0x0a, 0x04,
			0x65, 0x6e, 0x67, 0x00, 0x75, 0xff, 0x59, 0x3a}

		// Create expected PMT struct
		expectedProgDesc := make([]model.Descriptor, 0)

		videoStream := model.DataStream{StreamPid: 32, StreamType: 2, StreamDesc: make([]model.Descriptor, 0)}
		audioStream := model.DataStream{StreamPid: 33, StreamType: 4, StreamDesc: []model.Descriptor{{Tag: 10, Content: "65 6e 67 0"}}}
		streams := []model.DataStream{videoStream, audioStream}

		expectedPmt := model.CreatePMT(258, 2, 10, 0, true, expectedProgDesc, streams, -1)

		// Parse
		parsed := model.ParsePMT(dummyPMT, 258, 0)

		err := testUtils.Assert_obj_equal(expectedPmt, parsed)
		return nil, err
	})

	return tc
}

func readAdaptationFieldTest() testUtils.Testcase {
	tc := testUtils.GetTestCase("readAdaptationFieldTest", 0)

	tc.Describe("Initialization", func(input interface{}) (interface{}, error) {
		dummyAdaptationField := []byte{0x07, 0x50, 0x00, 0x04, 0xce, 0xcd, 0x7e, 0xf3}

		expected := model.AdaptationField{AfLen: 8, Pcr: 189051243, Splice_point: -1, Private_data: ""}
		parsed := model.ParseAdaptationField(dummyAdaptationField)

		err := testUtils.Assert_obj_equal(expected, parsed)
		if err != nil {
			return nil, err
		} else {
			return nil, nil
		}
	})

	return tc
}

func readPesHeaderTest() testUtils.Testcase {
	tc := testUtils.GetTestCase("readPesHeaderTest", 0)

	tc.Describe("Parse Pes header", func(input interface{}) (interface{}, error) {
		dummyPesHeader := []byte{0x00, 0x00, 0x01, 0xea, 0x7d, 0xb2,
			0x8f, 0xc0, 0x0a, 0x31, 0x00, 0x2b, 0x85, 0xfb,
			0x11, 0x00, 0x2b, 0x31, 0x9b}

		parsed, err := model.ParsePESHeader(dummyPesHeader)
		if err != nil {
			return nil, err
		}
		expected := model.CreatePESHeader(234, 32165, model.CreateOptionalPESHeader(13, 705277, 694477))

		assertErr := testUtils.Assert_obj_equal(expected, parsed)
		if assertErr != nil {
			return nil, assertErr
		} else {
			return nil, nil
		}
	})

	return tc
}

func readSCTE35SectionTest() testUtils.Testcase {
	tc := testUtils.GetTestCase(("readSCTE35SectionTest"), 0)

	tc.Describe("Read Splice insert", func(input interface{}) (interface{}, error) {
		dummySection := []byte{0x00, 0xfc, 0x30, 0x25, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0xff, 0xf0, 0x14, 0x05, 0x00, 0x00,
			0x00, 0x02, 0x7f, 0xef, 0xfe, 0x00, 0x2e, 0xb0, 0x30, 0xfe,
			0x00, 0x14, 0x99, 0x70, 0x00, 0x01, 0x00, 0x01, 0x00, 0x00,
			0xbb, 0x9e, 0x64, 0x39}

		parsed := model.ReadSCTE35Section(dummySection, 1)

		spliceInsert := model.Splice_event{EventId: 2, EventCancelIdr: false, OutOfNetworkIdr: true, ProgramSpliceFlag: true,
			DurationFlag: true, SpliceImmediateFlag: false, SpliceTime: 3059760, Components: []model.Splice_component{}, BreakDuration: model.Break_duration{AutoReturn: true, Duration: 1350000},
			UniqueProgramId: 1, AvailNum: 0, AvailsExpected: 1}

		expected := model.Splice_info_section{TableId: 252, SectionSyntaxIdr: false, PrivateIdr: false,
			SectionLen: 37, ProtocolVersion: 0, EncryptedPkt: false, EncryptAlgo: 0,
			PtsAdjustment: 0, CwIdx: 0, Tier: 4095, SpliceCmdLen: 20, SpliceCmdType: 5, SpliceSchedule: model.Splice_schedule{},
			SpliceInsert: spliceInsert, TimeSignal: model.Time_signal{}, PrivateCommand: model.Private_command{}}

		assertErr := testUtils.Assert_obj_equal(expected, parsed)
		if assertErr != nil {
			return nil, assertErr
		} else {
			return nil, nil
		}
	})

	return tc
}

func AddUnitTestSuite(t *testUtils.Tester) {
	tmg := testUtils.GetTestCaseMgr()

	// We may add custom test filter here later

	// model
	tmg.AddTest(readPATTest, []string{"avContainer"})
	tmg.AddTest(readPMTTest, []string{"avContainer"})
	tmg.AddTest(readAdaptationFieldTest, []string{"avContainer"})
	tmg.AddTest(readPesHeaderTest, []string{"avContainer"})
	tmg.AddTest(readSCTE35SectionTest, []string{"avContainer"})

	// demux
	tsdemux.AddUnitTestSuite(t)

	t.AddSuite("unitTest", tmg)
}
