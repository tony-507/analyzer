package ioUtils

import (
	"fmt"
	"os"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tony-507/analyzers/src/common"
	"github.com/tony-507/analyzers/src/testUtils"
)

// Helper
var TEST_OUT_DIR = testUtils.GetOutputDir() + "/test_output/"

func TestReaderSetParameter(t *testing.T) {
	specs := []IOReaderParam{
		IOReaderParam{Source: SOURCE_FILE, FileInput: FileInputParam{Fname: "dummy.ts"}},
		IOReaderParam{Source: SOURCE_FILE, FileInput: FileInputParam{Fname: "hello.abc.ts"}},
		IOReaderParam{Source: SOURCE_FILE, FileInput: FileInputParam{Fname: "hello.abc"}},
	}

	expectedExt := []INPUT_TYPE{INPUT_TS, INPUT_TS, INPUT_UNKNOWN}

	for idx, param := range specs {
		fr := GetReader("dummy")
		fr.SetParameter(param)

		impl, isFileReader := fr.impl.(*fileReader)
		if !isFileReader {
			panic("File reader not created")
		}
		assert.Equal(t, expectedExt[idx], impl.ext, fmt.Sprintf("[%d] Input file extension should be %v", idx, expectedExt[idx]))
		assert.Equal(t, -1, fr.maxInCnt, "Uninitialized maxInCnt should be -1")
	}
}

func TestReaderDeliverUnit(t *testing.T) {
	specs := []IOReaderParam{
		IOReaderParam{Source: SOURCE_DUMMY},
		IOReaderParam{Source: SOURCE_DUMMY, SkipCnt: 2},  // Deliver with skipping does not change behaviour
		IOReaderParam{Source: SOURCE_DUMMY, MaxInCnt: 2}, // Deliver with max input count
	}

	expectedDeliverCnt := []int{5, 5, 2}

	for idx, param := range specs {
		ir := GetReader("dummy")
		ir.SetParameter(param)

		ir.SetCallback(func(s string, reqType common.WORKER_REQUEST, obj interface{}) {
			expected := common.MakeReqUnit(ir.name, common.FETCH_REQUEST)
			assert.Equal(t, expected, obj, fmt.Sprintf("[%d] Expect a fetch request", idx))
		})

		for i := 0; i < expectedDeliverCnt[idx]; i++ {
			ir.start()
		}

		ir.SetCallback(func(s string, reqType common.WORKER_REQUEST, obj interface{}) {
			expected := common.MakeReqUnit(ir.name, common.EOS_REQUEST)
			assert.Equal(t, expected, obj, "Expect an EOS request")
		})
		ir.start()
	}
}

func TestOutputWriterDeliverUnit(t *testing.T) {
	ow := GetOutputWriter("dummy")
	x := 0
	m_parameter := IOWriterParam{OutputType: OUTPUT_DUMMY, dummyOut: &x}
	ow.SetParameter(m_parameter)

	// Deliver some dummy units
	for i := 1; i < 5; i++ {
		unit := common.IOUnit{Buf: i}
		ow.DeliverUnit(unit)
	}

	assert.Equal(t, 1234, x, "Expect output to be 1234")
}

func TestWriterMultiThread(t *testing.T) {
	fw := GetFileWriter()
	param := IOWriterParam{FileOutput: FileOutputParam{OutFolder: TEST_OUT_DIR}}
	fw.setup(param)

	buf1 := common.MakeSimpleBuf([]byte{})
	buf1.SetField("addPid", true, false)
	buf1.SetField("pid", 5, false)
	buf1.SetField("type", 3, false)

	buf2 := common.MakeSimpleBuf([]byte{})
	buf2.SetField("addPid", true, false)
	buf2.SetField("pid", 2, false)
	buf2.SetField("type", 3, false)

	fw.processControl(common.MakeStatusUnit(0x10, &buf1))
	fw.processControl(common.MakeStatusUnit(0x10, &buf2))

	rawUnit := common.IOUnit{Buf: []byte{1}, IoType: 3, Id: 5}
	rawUnit2 := common.IOUnit{Buf: []byte{1}, IoType: 3, Id: 2}
	fw.processUnit(rawUnit)
	fw.processUnit(rawUnit2)

	fw.stop()

	filesArr := []int{2, 5}
	for _, id := range filesArr {
		f, _ := os.Open(TEST_OUT_DIR + "out_" + strconv.Itoa(id) + ".ts")
		data := make([]byte, 1)
		f.Read(data)
		assert.Equal(t, uint8(1), data[0], "Expect 1")
	}
}
