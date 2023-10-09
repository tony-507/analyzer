package dataHandler

import (
	"sort"

	"github.com/tony-507/analyzers/src/common"
	"github.com/tony-507/analyzers/src/plugins/dataHandler/utils"
)

type videoDataProcessorStruct struct {
	videos []utils.VideoDataStruct
	logger common.Log
}

func (vp *videoDataProcessorStruct) process(cmBuf common.CmBuf, data *utils.VideoDataStruct) {
	dts, _ := common.GetBufFieldAsInt(cmBuf, "dts")
	pts, _ := common.GetBufFieldAsInt(cmBuf, "pts")
	data.Dts = dts
	data.Pts = pts

	if len(vp.videos) == 20 {
		// Clear and display stored data
		for _, storedData := range vp.videos[:10] {
			if storedData.TimeCode.Frame != -1 {
				vp.logger.Info("Timecode received at PTS %d -- %s", storedData.Pts, storedData.TimeCode.ToString())
			}
		}
		vp.videos = vp.videos[10:]
	}

	vp.videos = append(vp.videos, *data)
	sort.Slice(vp.videos, func (i, j int) bool { return vp.videos[i].Pts < vp.videos[j].Pts })
}

func videoDataProcessor() videoDataProcessorStruct {
	return videoDataProcessorStruct{
		videos: make([]utils.VideoDataStruct, 0, 20),
		logger: common.CreateLogger("VideoDataProcessor"),
	}
}
