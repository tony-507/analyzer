package dataHandler

import (
	"fmt"
	"sort"
	"strings"

	"github.com/tony-507/analyzers/src/logging"
	"github.com/tony-507/analyzers/src/plugins/common"
	"github.com/tony-507/analyzers/src/plugins/common/io"
	"github.com/tony-507/analyzers/src/plugins/dataHandler/utils"
	"github.com/tony-507/analyzers/src/tttKernel"
)

type processorStat struct {
	nTcDiscontinuity int
}

type videoDataProcessorStruct struct {
	videos    []utils.VideoDataStruct
	splicePTS []uint64
	lastTC    common.TimeCode
	writer    io.FileWriter
	logger    logging.Log
	stat      processorStat
}

func (vp *videoDataProcessorStruct) Start() error {
	return vp.writer.Open()
}

func (vp *videoDataProcessorStruct) Stop() error {
	return vp.writer.Close()
}

func (vp *videoDataProcessorStruct) Process(unit tttKernel.CmUnit, parsedData *utils.ParsedData) {
	cmBuf := unit.GetBuf()

	switch parsedData.GetType() {
	case utils.PARSED_VIDEO:
		vUnit, _ := unit.(*common.MediaUnit)
		vmd := vUnit.GetVideoData()

		data := parsedData.GetVideoData()
		dts, _ := tttKernel.GetBufFieldAsInt(cmBuf, "dts")
		pts, _ := tttKernel.GetBufFieldAsInt(cmBuf, "pts")
		data.Dts = dts
		data.Pts = pts

		if data.Type == common.I_SLICE || data.Type == common.IDR_SLICE {
			// Sort data when we reach an I slice
			sort.Slice(vp.videos, func (i, j int) bool { return vp.videos[i].Pts < vp.videos[j].Pts })
			for _, storedData := range vp.videos {
				vp.writer.Write(storedData.ToCmBuf())

				vp.validateTimeCode(&storedData)
				vp.validateSpliceIDR(&storedData)
			}
			vp.videos = []utils.VideoDataStruct{}
		}

		vmd.Type = common.FRAME_TYPE(data.Type)
		vmd.Tc = data.TimeCode

		vp.videos = append(vp.videos, *data)
	case utils.PARSED_DATA:
		data := parsedData.GetData()
		if data.Type == utils.SCTE_35 {
			vp.splicePTS = append(vp.splicePTS, uint64(data.Scte35.SpliceTime))
		}
	}
}

func (vp *videoDataProcessorStruct) validateTimeCode(data *utils.VideoDataStruct) {
	if !data.TimeCode.IsEmpty() {
		// Currently assume 29.97 with drop frame
		nextTc := common.GetNextTimeCode(&vp.lastTC, 30000, 1001, true)
		if !data.TimeCode.Equals(&nextTc) {
			vp.stat.nTcDiscontinuity++
		}
		vp.lastTC = data.TimeCode
	}
}

func (vp *videoDataProcessorStruct) validateSpliceIDR(data *utils.VideoDataStruct) {
	if data.Type != common.I_SLICE {
		return
	}

	updatedSplicePTS := []uint64{}
	for _, spliceTime := range vp.splicePTS {
		if uint64(data.Pts) > spliceTime {
			vp.logger.Error("No I frame at splice PTS %d", spliceTime)
		} else if uint64(data.Pts) == spliceTime {
			vp.logger.Info("I frame found for splice PTS %d", spliceTime)
		} else {
			updatedSplicePTS = append(updatedSplicePTS, spliceTime)
		}
	}
	vp.splicePTS = updatedSplicePTS
}

func (vp *videoDataProcessorStruct) PrintInfo(sb *strings.Builder) {
	sb.WriteString("\tVideo processor:\n")
	if !vp.lastTC.IsEmpty() {
		sb.WriteString(fmt.Sprintf("\t\tTimecode discontinuities: %d", vp.stat.nTcDiscontinuity))
		vp.stat.nTcDiscontinuity = 0
	}
}

func videoDataProcessor(outDir string) utils.DataProcessor {
	return &videoDataProcessorStruct{
		videos: make([]utils.VideoDataStruct, 0, 20),
		lastTC: common.NewTimeCode(),
		writer: io.CsvWriter(outDir, "video.csv"),
		logger: logging.CreateLogger("VideoDataProcessor"),
		stat: processorStat{},
	}
}
