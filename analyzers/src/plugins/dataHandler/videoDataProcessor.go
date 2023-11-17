package dataHandler

import (
	"sort"
	"strings"

	"github.com/tony-507/analyzers/src/common"
	"github.com/tony-507/analyzers/src/common/logging"
	"github.com/tony-507/analyzers/src/plugins/dataHandler/utils"
	commonUtils "github.com/tony-507/analyzers/src/utils"
)

type videoDataProcessorStruct struct {
	videos   []utils.VideoDataStruct
	lastTC   common.TimeCode
	tcWriter commonUtils.FileWriter
	logger   logging.Log
}

func (vp *videoDataProcessorStruct) Start() error {
	return vp.tcWriter.Open()
}

func (vp *videoDataProcessorStruct) Stop() error {
	return vp.tcWriter.Close()
}

func (vp *videoDataProcessorStruct) Process(unit common.CmUnit, parsedData *utils.ParsedData) {
	if parsedData.GetType() != utils.PARSED_VIDEO {
		return
	}
	cmBuf := unit.GetBuf()
	vUnit, _ := unit.(*common.MediaUnit)
	vmd := vUnit.GetVideoData()

	data := parsedData.GetVideoData()
	dts, _ := common.GetBufFieldAsInt(cmBuf, "dts")
	pts, _ := common.GetBufFieldAsInt(cmBuf, "pts")
	data.Dts = dts
	data.Pts = pts

	if data.Type == utils.I_SLICE {
		// Sort data when we reach an I slice
		sort.Slice(vp.videos, func (i, j int) bool { return vp.videos[i].Pts < vp.videos[j].Pts })
		for _, storedData := range vp.videos {
			if !storedData.TimeCode.IsEmpty() {
				// Currently assume 29.97 with drop frame
				nextTc := commonUtils.GetNextTimeCode(&vp.lastTC, 30000, 1001, true)
				if !storedData.TimeCode.Equals(&nextTc) {
					vp.logger.Error("VITC jump detected: %s -> %s. Expecting %s", vp.lastTC.ToString(), storedData.TimeCode.ToString(), nextTc.ToString())
				}
				vp.tcWriter.Write(storedData.ToCmBuf())
				vp.lastTC = storedData.TimeCode
			}
		}
		vp.videos = []utils.VideoDataStruct{}
	}

	vmd.Type = common.FRAME_TYPE(data.Type)
	vmd.Tc = data.TimeCode

	vp.videos = append(vp.videos, *data)
}

func (vp *videoDataProcessorStruct) PrintInfo(sb *strings.Builder) {}

func videoDataProcessor() utils.DataProcessor {
	return &videoDataProcessorStruct{
		videos: make([]utils.VideoDataStruct, 0, 20),
		lastTC: common.NewTimeCode(),
		tcWriter: commonUtils.CsvWriter("output", "vitc.csv"),
		logger: logging.CreateLogger("VideoDataProcessor"),
	}
}
