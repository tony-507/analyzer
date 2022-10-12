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

func readAdaptationFieldTest() testUtils.Testcase {
	tc := testUtils.GetTestCase("readAdaptationFieldTest")

	tc.Describe("Initialization", func(input interface{}) (interface{}, error) {
		dummyAdaptationField := []byte{0x07, 0x50, 0x00, 0x04, 0xce, 0xcd, 0x7e, 0xf3}
		return dummyAdaptationField, nil
	})

	tc.Describe("Parse adaptation field", func(input interface{}) (interface{}, error) {
		dummyAdaptationField, isBts := input.([]byte)
		if !isBts {
			err := errors.New("Input not passed to next step")
			return nil, err
		}

		expected := AdaptationField{afLen: 8, pcr: 189051243, splice_point: -1, private_data: ""}
		parsed := ParseAdaptationField(dummyAdaptationField)

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
	tc := testUtils.GetTestCase("readPesHeaderTest")

	tc.Describe("Parse Pes header", func(input interface{}) (interface{}, error) {
		dummyPesHeader := []byte{0x00, 0x00, 0x01, 0xea, 0x7d, 0xb2,
			0x8f, 0xc0, 0x0a, 0x31, 0x00, 0x2b, 0x85, 0xfb,
			0x11, 0x00, 0x2b, 0x31, 0x9b}

		parsed, err := ParsePESHeader(dummyPesHeader)
		if err != nil {
			return nil, err
		}
		expected := PESHeader{streamId: 234, sectionLen: 32165, optionalHeader: OptionalHeader{scrambled: false,
			dataAligned: true, length: 13, pts: 705277, dts: 694477}}

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
	tc := testUtils.GetTestCase(("readSCTE35SectionTest"))

	tc.Describe("Read Splice insert", func(input interface{}) (interface{}, error) {
		dummySection := []byte{0x00, 0xfc, 0x30, 0x25, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0xff, 0xf0, 0x14, 0x05, 0x00, 0x00,
			0x00, 0x02, 0x7f, 0xef, 0xfe, 0x00, 0x2e, 0xb0, 0x30, 0xfe,
			0x00, 0x14, 0x99, 0x70, 0x00, 0x01, 0x00, 0x01, 0x00, 0x00,
			0xbb, 0x9e, 0x64, 0x39}

		parsed := readSCTE35Section(dummySection, 1)

		spliceInsert := splice_event{EventId: 2, EventCancelIdr: false, OutOfNetworkIdr: true, ProgramSpliceFlag: true,
			DurationFlag: true, SpliceImmediateFlag: false, SpliceTime: 3059760, Components: []splice_component{}, BreakDuration: break_duration{AutoReturn: true, Duration: 1350000},
			UniqueProgramId: 1, AvailNum: 0, AvailsExpected: 1}

		expected := splice_info_section{TableId: 252, SectionSyntaxIdr: false, privateIdr: false,
			SectionLen: 37, ProtocolVersion: 0, EncryptedPkt: false, EncryptAlgo: 0,
			PtsAdjustment: 0, CwIdx: 0, Tier: 4095, SpliceCmdLen: 20, SpliceCmdType: 5, SpliceSchedule: splice_schedule{},
			SpliceInsert: spliceInsert, TimeSignal: time_signal{}, PrivateCommand: private_command{}}

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

	// common
	tmg.AddTest(readPATTest, []string{"avContainer"})
	tmg.AddTest(readPMTTest, []string{"avContainer"})
	tmg.AddTest(readAdaptationFieldTest, []string{"avContainer"})
	tmg.AddTest(readPesHeaderTest, []string{"avContainer"})
	tmg.AddTest(readSCTE35SectionTest, []string{"avContainer"})

	t.AddSuite("unitTest", tmg)
}
