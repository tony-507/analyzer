package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type dummyManagerStruct struct {
	programRecords map[int]int
	patVersion     int
	psiJsons       map[int][]byte
}

func (m *dummyManagerStruct) AddProgram(version int, progNum int, pid int) {
	m.programRecords[progNum] = pid
	m.patVersion = version
}

func (m *dummyManagerStruct) GetPATVersion() int {
	return m.patVersion
}

func (m *dummyManagerStruct) PsiUpdateFinished(pid int, jsonBytes []byte) {
	m.psiJsons[pid] = jsonBytes
}

func dummyManager() *dummyManagerStruct {
	return &dummyManagerStruct{programRecords: make(map[int]int, 0), patVersion: -1, psiJsons: make(map[int][]byte, 0)}
}

func TestReadPAT(t *testing.T) {
	dummyPAT := []byte{0x00, 0x00, 0xB0, 0x0D, 0x11, 0x11, 0xC1,
		0x00, 0x00, 0x00, 0x0A, 0xE1, 0x02, 0xAA, 0x4A, 0xE2, 0xD2}
	manager := dummyManager()

	table, err := PsiTable(manager, 0, dummyPAT)
	if err != nil {
		panic(err)
	}
	parseErr := table.ParsePayload()
	if parseErr != nil {
		panic(parseErr)
	}
	assert.Equal(t, true, table.Ready(), "PAT should be ready for parsing")
	assert.Equal(t, []byte{0x7b, 0xa, 0x9, 0x9, 0x22, 0x50, 0x6b, 0x74, 0x43,
		0x6e, 0x74, 0x22, 0x3a, 0x20, 0x30, 0x2c, 0xa, 0x9, 0x9, 0x22, 0x56, 0x65,
		0x72, 0x73, 0x69, 0x6f, 0x6e, 0x22, 0x3a, 0x20, 0x30, 0x2c, 0xa, 0x9, 0x9,
		0x22, 0x50, 0x72, 0x6f, 0x67, 0x72, 0x61, 0x6d, 0x4d, 0x61, 0x70, 0x22, 0x3a,
		0x20, 0x7b, 0xa, 0x9, 0x9, 0x9, 0x22, 0x31, 0x30, 0x22, 0x3a, 0x20, 0x32, 0x35,
		0x38, 0xa, 0x9, 0x9, 0x7d, 0x2c, 0xa, 0x9, 0x9, 0x22, 0x43, 0x72, 0x63, 0x33,
		0x32, 0x22, 0x3a, 0x20, 0x31, 0x30, 0xa, 0x9, 0x7d},
		manager.psiJsons[0], "PAT content not match")
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
		{0x00},
		{0x07, 0x50, 0x00, 0x04, 0xce, 0xcd, 0x7e, 0xf3}, // With PCR
		{0x14, 0x5e, 0x00, 0x04, 0xce, 0xcd, 0x7e, 0xf3, 0x00, 0x04, 0xce, 0xcd, 0x7e, 0xf3, 0x01, 0x03, 0x45, 0x4e, 0x47, 0xff, 0xff}, // With everything
	}
	structSpecs := []AdaptationField{
		{AfLen: 0},
		{AfLen: 7, DisCnt_cnt: 0, RandomAccess: 1, EsIdr: 0, Pcr: 189051243, Opcr: -1, Splice_point: -1, Private_data: "", StuffSize: 0},
		{AfLen: 20, DisCnt_cnt: 0, RandomAccess: 1, EsIdr: 0, Pcr: 189051243, Opcr: 189051243, Splice_point: 1, Private_data: "ENG", StuffSize: 2},
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
		{0x00, 0x00, 0x01, 0xea, 0x17, 0xb2, 0x8f, 0x00, 0x00},
		{0x00, 0x00, 0x01, 0xea, 0x17, 0xb2, 0x8f, 0x80, 0x05, 0x21, 0x00, 0x2b, 0x4d, 0xbb},
		{0x00, 0x00, 0x01, 0xea, 0x7d, 0xb2, 0x8f, 0xc0, 0x0a, 0x31, 0x00, 0x2b, 0x85, 0xfb, 0x11, 0x00, 0x2b, 0x31, 0x9b},
		{0x00, 0x00, 0x01, 0xea, 0x00, 0x00, 0x8f, 0x00, 0x00, 0x01, 0x02, 0x03, 0x04, 0x05},
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
		{0x00, 0xfc, 0x30, 0x11, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xff, 0xf0, 0x00, 0x00, 0x00, 0x00}, // Splice null
		{0x00, 0xfc, 0x30, 0x25, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0xff, 0xf0, 0x14, 0x05, 0x00, 0x00, 0x00, 0x02, 0x7f,
			0xef, 0xfe, 0x00, 0x2e, 0xb0, 0x30, 0xfe, 0x00, 0x14, 0x99,
			0x70, 0x00, 0x01, 0x00, 0x01, 0x00, 0x00, 0xbb, 0x9e, 0x64, 0x39}, // Splice insert
	}
	structSpecs := []Splice_info_section{
		{TableId: 252, SectionSyntaxIdr: false, PrivateIdr: false,
			SectionLen: 17, ProtocolVersion: 0, EncryptedPkt: false, EncryptAlgo: 0,
			PtsAdjustment: 0, CwIdx: 0, Tier: 4095, SpliceCmdLen: 0, SpliceCmdType: 0, SpliceCmdTypeStr: "splice_null",
			SpliceCmd: Splice_null{},
		},
		{TableId: 252, SectionSyntaxIdr: false, PrivateIdr: false,
			SectionLen: 37, ProtocolVersion: 0, EncryptedPkt: false, EncryptAlgo: 0,
			PtsAdjustment: 0, CwIdx: 0, Tier: 4095, SpliceCmdLen: 20, SpliceCmdType: 5, SpliceCmdTypeStr: "splice_insert",
			SpliceCmd: Splice_event{EventId: 2, EventCancelIdr: false, OutOfNetworkIdr: true, ProgramSpliceFlag: true,
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
	rawTsPacket := append(headerBytes, make([]byte, 184)...)

	pkt, err := TsPacket(rawTsPacket)
	if err != nil {
		panic(err)
	}

	tei, fieldErr := pkt.GetField("tei")
	if fieldErr != nil {
		panic(fieldErr)
	}
	assert.Equal(t, 0, tei, "tei not match")

	pusi, fieldErr := pkt.GetField("pusi")
	if fieldErr != nil {
		panic(fieldErr)
	}
	assert.Equal(t, 0, pusi, "pusi not match")

	priority, fieldErr := pkt.GetField("priority")
	if fieldErr != nil {
		panic(fieldErr)
	}
	assert.Equal(t, 0, priority, "priority not match")

	pid, fieldErr := pkt.GetField("pid")
	if fieldErr != nil {
		panic(fieldErr)
	}
	assert.Equal(t, 911, pid, "pid not match")

	tsc, fieldErr := pkt.GetField("tsc")
	if fieldErr != nil {
		panic(fieldErr)
	}
	assert.Equal(t, 0, tsc, "tsc not match")

	afc, fieldErr := pkt.GetField("afc")
	if fieldErr != nil {
		panic(fieldErr)
	}
	assert.Equal(t, 1, afc, "afc not match")

	cc, fieldErr := pkt.GetField("cc")
	if fieldErr != nil {
		panic(fieldErr)
	}
	assert.Equal(t, 15, cc, "cc not match")

	// headerStruct := TsHeader{Tei: false, Pusi: false, Priority: false, Pid: 911, Tsc: 0, Afc: 1, Cc: 15}
}
