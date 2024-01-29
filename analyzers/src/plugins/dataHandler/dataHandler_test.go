package dataHandler

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tony-507/analyzers/src/plugins/common"
	"github.com/tony-507/analyzers/src/plugins/dataHandler/utils"
)

func TestScte35IDR(t *testing.T) {
	/*
	 * When IDR match is found, return true
	 * When no IDR match is found, return false
	 * Drop all expired SCTE-35 splice time
	 * Should work under splice repetition
	 */
	proc, _ := videoDataProcessor("dummy").(*videoDataProcessorStruct)
	spliceTimes := []uint64{2233567, 3344567}
	idrPts := []uint64{1234567, 2233567, 3344577}
	expected := []bool{true, true, false}
	remaining := []int{2, 1, 0}

	for _, spliceTime := range spliceTimes {
		newData := utils.CreateParsedData()
		data := newData.GetData()
		data.Type = utils.SCTE_35
		data.Scte35 = &utils.Scte35Struct{
			SpliceTime: int(spliceTime),
		}
		proc.Process(nil, &newData)
		proc.Process(nil, &newData)
	}

	for idx, pts := range idrPts {
		newData := utils.CreateParsedData()
		data := newData.GetVideoData()
		data.Type = common.IDR_SLICE
		data.Pts = int(pts)
		data.Dts = int(pts)
		assert.Equal(t, expected[idx], proc.validateSpliceIDR(data), "Splice IDR validation returns wrong result")
		assert.Equal(t, remaining[idx], len(proc.splicePTS), "Splice PTS not dropped when expired")
	}
}
