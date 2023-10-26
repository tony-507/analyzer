package dataHandler

import (
	"sort"
	"strings"

	"github.com/tony-507/analyzers/src/common"
	"github.com/tony-507/analyzers/src/plugins/dataHandler/utils"
	commonUtils "github.com/tony-507/analyzers/src/utils"
)

type videoDataProcessorStruct struct {
	videos   []utils.VideoDataStruct
	lastTC   commonUtils.TimeCode
	tcWriter commonUtils.FileWriter
	logger   common.Log
}

func (vp *videoDataProcessorStruct) Start() error {
	return vp.tcWriter.Open()
}

func (vp *videoDataProcessorStruct) Stop() error {
	return vp.tcWriter.Close()
}

func (vp *videoDataProcessorStruct) Process(cmBuf common.CmBuf, parsedData *utils.ParsedData) {
	if parsedData.GetType() != utils.PARSED_VIDEO {
		return
	}

	data := parsedData.GetVideoData()
	dts, _ := common.GetBufFieldAsInt(cmBuf, "dts")
	pts, _ := common.GetBufFieldAsInt(cmBuf, "pts")
	data.Dts = dts
	data.Pts = pts

	if len(vp.videos) == 20 {
		// Clear and display stored data
		sort.Slice(vp.videos, func (i, j int) bool { return vp.videos[i].Pts < vp.videos[j].Pts })
		for _, storedData := range vp.videos[:10] {
			if storedData.TimeCode.Frame != -1 {
				// Currently assume 29.97 with drop frame
				nextTc := commonUtils.GetNextTimeCode(&vp.lastTC, 30000, 1001, true)
				if !storedData.TimeCode.Equals(&nextTc) {
					vp.logger.Error("VITC jump detected: %s -> %s. Expecting %s", vp.lastTC.ToString(), storedData.TimeCode.ToString(), nextTc.ToString())
				}
				vp.tcWriter.Write(storedData.ToCmBuf())
				vp.lastTC = storedData.TimeCode
			}
		}
		vp.videos = vp.videos[10:]
	}

	vp.videos = append(vp.videos, *data)
}

func (vp *videoDataProcessorStruct) PrintInfo(sb *strings.Builder) {}

func videoDataProcessor() utils.DataProcessor {
	return &videoDataProcessorStruct{
		videos: make([]utils.VideoDataStruct, 0, 20),
		lastTC: commonUtils.TimeCode{Frame: -1},
		tcWriter: commonUtils.CsvWriter("output", "vitc.csv"),
		logger: common.CreateLogger("VideoDataProcessor"),
	}
}
