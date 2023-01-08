package ioUtils

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tony-507/analyzers/src/common"
	"github.com/tony-507/analyzers/src/logs"
)

func getOutputDir() string {
	_, filename, _, _ := runtime.Caller(0)
	return filepath.Dir(filename)
}

// Helper
var TEST_OUT_DIR = getOutputDir() + "/../../../build/test_output/"

func TestReaderSetParameter(t *testing.T) {
	specs := []string{
		"{\"Source\":\"_SOURCE_FILE\",\"FileInput\":{\"Fname\":\"dummy.ts\"}}",
		"{\"Source\":\"_SOURCE_FILE\",\"FileInput\":{\"Fname\":\"hello.abc.ts\"}}",
		"{\"Source\":\"_SOURCE_FILE\",\"FileInput\":{\"Fname\":\"hello.abc\"}}",
	}

	expectedExt := []INPUT_TYPE{INPUT_TS, INPUT_TS, INPUT_UNKNOWN}

	for idx, param := range specs {
		fr := InputReader{name: "dummy", logger: logs.CreateLogger("dummy")}
		fr.setParameter(param)

		impl, isFileReader := fr.impl.(*fileReader)
		if !isFileReader {
			panic("File reader not created")
		}
		assert.Equal(t, expectedExt[idx], impl.ext, fmt.Sprintf("[%d] Input file extension should be %v", idx, expectedExt[idx]))
		assert.Equal(t, -1, fr.maxInCnt, "Uninitialized maxInCnt should be -1")
	}
}

func TestReaderDeliverUnit(t *testing.T) {
	specs := []string{
		"{\"Source\": \"_SOURCE_DUMMY\"}",
		"{\"Source\": \"_SOURCE_DUMMY\",\"SkipCnt\":2}",  // Deliver with skipping does not change behaviour
		"{\"Source\": \"_SOURCE_DUMMY\",\"MaxInCnt\":2}", // Deliver with max input count
	}

	expectedDeliverCnt := []int{5, 5, 2}

	for idx, param := range specs {
		ir := InputReader{name: "dummy", logger: logs.CreateLogger("dummy")}
		ir.setParameter(param)

		ir.setCallback(func(s string, reqType common.WORKER_REQUEST, obj interface{}) {
			expected := common.MakeReqUnit(ir.name, common.FETCH_REQUEST)
			assert.Equal(t, expected, obj, fmt.Sprintf("[%d] Expect a fetch request", idx))
		})

		for i := 0; i < expectedDeliverCnt[idx]; i++ {
			ir.start()
		}

		ir.setCallback(func(s string, reqType common.WORKER_REQUEST, obj interface{}) {
			expected := common.MakeReqUnit(ir.name, common.EOS_REQUEST)
			assert.Equal(t, expected, obj, "Expect an EOS request")
		})
		ir.start()
	}
}

func TestWriterMultiThread(t *testing.T) {
	fw := getFileWriter()
	param := ioWriterParam{FileOutput: fileOutputParam{OutFolder: TEST_OUT_DIR}}
	fw.setup(param)

	buf1 := common.MakeSimpleBuf([]byte{})
	buf1.SetField("addPid", true, false)
	buf1.SetField("pid", 5, false)
	buf1.SetField("type", 3, false)

	buf2 := common.MakeSimpleBuf([]byte{})
	buf2.SetField("addPid", true, false)
	buf2.SetField("pid", 2, false)
	buf2.SetField("type", 3, false)

	fw.processControl(common.MakeStatusUnit(0x10, buf1))
	fw.processControl(common.MakeStatusUnit(0x10, buf2))

	rawUnit := common.MakeIOUnit([]byte{1}, 3, 5)
	rawUnit2 := common.MakeIOUnit([]byte{1}, 3, 2)
	fw.processUnit(rawUnit)
	fw.processUnit(rawUnit2)

	fw.stop()

	t.Logf(TEST_OUT_DIR)

	filesArr := []int{2, 5}
	for _, id := range filesArr {
		f, _ := os.Open(TEST_OUT_DIR + "out_" + strconv.Itoa(id) + ".ts")
		data := make([]byte, 1)
		f.Read(data)
		assert.Equal(t, uint8(1), data[0], "Expect 1")
	}
}
