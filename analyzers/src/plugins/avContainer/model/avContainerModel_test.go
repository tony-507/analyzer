package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadPAT(t *testing.T) {
	dummyPAT := []byte{0x00, 0x00, 0xB0, 0x0D, 0x11, 0x11, 0xC1,
		0x00, 0x00, 0x00, 0x0A, 0xE1, 0x02, 0xAA, 0x4A, 0xE2, 0xD2}

	assert.Equal(t, true, PATReadyForParse(dummyPAT), "PAT should be ready for parsing")

	PAT, parsingErr := ParsePAT(dummyPAT, 0)
	if parsingErr != nil {
		panic(parsingErr)
	}

	programMap := make(map[int]int, 0)
	programMap[258] = 10
	expected := CreatePAT(0, 4369, 0, true, programMap, 10)

	assert.Equal(t, expected, PAT, "PAT not match")
	assert.Equal(t, "PktCnt 0 tableId 0 tableIdExt 4369 version 0 curNextIdr true crc 10\npid 258 => progNum 10", PAT.ToString(), "PAT string not match")
}

func TestReadPMT(t *testing.T) {
	dummyPMT := []byte{0x00, 0x02, 0xb0, 0x1d, 0x00, 0x0a, 0xc1,
		0x00, 0x00, 0xe0, 0x20, 0xf0, 0x00, 0x02, 0xe0, 0x20,
		0xf0, 0x00, 0x04, 0xe0, 0x21, 0xf0, 0x06, 0x0a, 0x04,
		0x65, 0x6e, 0x67, 0x00, 0x75, 0xff, 0x59, 0x3a}

	assert.Equal(t, true, PMTReadyForParse(dummyPMT), "PMT should be ready for parsing")

	// Create expected PMT struct
	expectedProgDesc := make([]Descriptor, 0)

	videoStream := DataStream{StreamPid: 32, StreamType: 2, StreamDesc: make([]Descriptor, 0)}
	audioStream := DataStream{StreamPid: 33, StreamType: 4, StreamDesc: []Descriptor{{Tag: 10, Content: "65 6e 67 00"}}}
	streams := []DataStream{videoStream, audioStream}

	expectedPmt := CreatePMT(258, 2, 10, 0, true, expectedProgDesc, streams, -1)

	// Parse
	parsed := ParsePMT(dummyPMT, 258, 0)

	assert.Equal(t, expectedPmt, parsed, "PMT not match")
}

func TestAdaptationFieldIO(t *testing.T) {
	caseName := []string{
		"EmptyAdaptationField",
		"AdapationFieldWithPCR",
		"AdapationFieldWithEverything",
	}
	byteSpecs := [][]byte{
		[]byte{0x00},
		[]byte{0x07, 0x50, 0x00, 0x04, 0xce, 0xcd, 0x7e, 0xf3},                                                                               // With PCR
		[]byte{0x14, 0x5e, 0x00, 0x04, 0xce, 0xcd, 0x7e, 0xf3, 0x00, 0x04, 0xce, 0xcd, 0x7e, 0xf3, 0x01, 0x03, 0x45, 0x4e, 0x47, 0xff, 0xff}, // With everything
	}
	structSpecs := []AdaptationField{
		AdaptationField{AfLen: 0},
		AdaptationField{AfLen: 7, DisCnt_cnt: 0, RandomAccess: 1, EsIdr: 0, Pcr: 189051243, Opcr: -1, Splice_point: -1, Private_data: "", StuffSize: 0},
		AdaptationField{AfLen: 20, DisCnt_cnt: 0, RandomAccess: 1, EsIdr: 0, Pcr: 189051243, Opcr: 189051243, Splice_point: 1, Private_data: "ENG", StuffSize: 2},
	}

	for idx := range byteSpecs {
		t.Run(caseName[idx], func(t *testing.T) {
			parsed := ParseAdaptationField(byteSpecs[idx])
			assert.Equal(t, structSpecs[idx], parsed, "Adaptation field struct not match")

			buf := structSpecs[idx].Serialize()
			assert.Equal(t, byteSpecs[idx], buf, "Adaptation field bytes not match")
		})
	}
}

func TestPesHeaderIO(t *testing.T) {
	caseName := []string{
		"PesHeaderWithNoTimestamp",
		"PesHeaderWithEqualTimestamp",
		"PesHeaderWithDiffTimestamp",
		"PesHeaderWithZeroLength",
	}
	byteSpecs := [][]byte{
		[]byte{0x00, 0x00, 0x01, 0xea, 0x17, 0xb2, 0x8f, 0x00, 0x00},
		[]byte{0x00, 0x00, 0x01, 0xea, 0x17, 0xb2, 0x8f, 0x80, 0x05, 0x21, 0x00, 0x2b, 0x4d, 0xbb},
		[]byte{0x00, 0x00, 0x01, 0xea, 0x7d, 0xb2, 0x8f, 0xc0, 0x0a, 0x31, 0x00, 0x2b, 0x85, 0xfb, 0x11, 0x00, 0x2b, 0x31, 0x9b},
		[]byte{0x00, 0x00, 0x01, 0xea, 0x00, 0x00, 0x8f, 0x00, 0x00, 0x01, 0x02, 0x03, 0x04, 0x05},
	}
	structSpecs := []PESHeader{
		CreatePESHeader(234, 6063, CreateOptionalPESHeader(3, -1, -1)),
		CreatePESHeader(234, 6058, CreateOptionalPESHeader(8, 698077, 698077)),
		CreatePESHeader(234, 32165, CreateOptionalPESHeader(13, 705277, 694477)),
		CreatePESHeader(234, 5, CreateOptionalPESHeader(3, -1, -1)),
	}

	for idx := range byteSpecs {
		t.Run(caseName[idx], func(t *testing.T) {
			parsed, _, err := ParsePESHeader(byteSpecs[idx])
			if err != nil {
				panic(err)
			}

			assert.Equal(t, structSpecs[idx], parsed, "PES header not match")
		})
	}
}

func TestSCTE35IO(t *testing.T) {
	caseName := []string{
		"SpliceNull",
		"SpliceInsert",
	}
	byteSpecs := [][]byte{
		[]byte{0x00, 0xfc, 0x30, 0x11, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xff, 0xf0, 0x00, 0x00, 0x00, 0x00}, // Splice null
		[]byte{0x00, 0xfc, 0x30, 0x25, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0xff, 0xf0, 0x14, 0x05, 0x00, 0x00, 0x00, 0x02, 0x7f,
			0xef, 0xfe, 0x00, 0x2e, 0xb0, 0x30, 0xfe, 0x00, 0x14, 0x99,
			0x70, 0x00, 0x01, 0x00, 0x01, 0x00, 0x00, 0xbb, 0x9e, 0x64, 0x39}, // Splice insert
	}
	structSpecs := []Splice_info_section{
		Splice_info_section{TableId: 252, SectionSyntaxIdr: false, PrivateIdr: false,
			SectionLen: 17, ProtocolVersion: 0, EncryptedPkt: false, EncryptAlgo: 0,
			PtsAdjustment: 0, CwIdx: 0, Tier: 4095, SpliceCmdLen: 0, SpliceCmdType: 0,
			Splice_command: Splice_null{},
		},
		Splice_info_section{TableId: 252, SectionSyntaxIdr: false, PrivateIdr: false,
			SectionLen: 37, ProtocolVersion: 0, EncryptedPkt: false, EncryptAlgo: 0,
			PtsAdjustment: 0, CwIdx: 0, Tier: 4095, SpliceCmdLen: 20, SpliceCmdType: 5,
			Splice_command: Splice_event{EventId: 2, EventCancelIdr: false, OutOfNetworkIdr: true, ProgramSpliceFlag: true,
				DurationFlag: true, SpliceImmediateFlag: false, SpliceTime: 3059760, Components: []Splice_component{},
				BreakDuration: Break_duration{AutoReturn: true, Duration: 1350000}, UniqueProgramId: 1, AvailNum: 0, AvailsExpected: 1},
		},
	}

	for idx := range byteSpecs {
		t.Run(caseName[idx], func(t *testing.T) {
			parsed := ReadSCTE35Section(byteSpecs[idx], 1)

			assert.Equal(t, structSpecs[idx], parsed, "SCTE-35 section not match")
		})
	}
}

func TestTsHeaderIO(t *testing.T) {
	headerBytes := []byte{0x47, 0x03, 0x8f, 0x1f}
	headerStruct := TsHeader{Tei: false, Pusi: false, Priority: false, Pid: 911, Tsc: 0, Afc: 1, Cc: 15}

	assert.Equal(t, headerStruct, ReadTsHeader(headerBytes), "TS header struct not match")
	assert.Equal(t, headerBytes, headerStruct.Serialize(), "TS header bytes not match")
}
