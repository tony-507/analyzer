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
	switch parsedData.GetType() {
	case utils.PARSED_VIDEO:
		cmBuf := unit.GetBuf()
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

				if !vp.validateTimeCode(&storedData) {
					vp.stat.nTcDiscontinuity++
				}
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
			hasSpliceTime := false
			for _, spliceTime := range vp.splicePTS {
				if spliceTime == uint64(data.Scte35.SpliceTime) {
					hasSpliceTime = true
					break
				}
			}
			if !hasSpliceTime {
				vp.splicePTS = append(vp.splicePTS, uint64(data.Scte35.SpliceTime))
			}
		}
	}
}

func (vp *videoDataProcessorStruct) validateTimeCode(data *utils.VideoDataStruct) bool {
	if !data.TimeCode.IsEmpty() {
		// Currently assume 29.97 with drop frame
		nextTc := common.GetNextTimeCode(&vp.lastTC, 30000, 1001, true)
		if !data.TimeCode.Equals(&nextTc) {
			return false
		}
		vp.lastTC = data.TimeCode
	}
	return true
}

func (vp *videoDataProcessorStruct) validateSpliceIDR(data *utils.VideoDataStruct) bool {
	if data.Type != common.IDR_SLICE {
		return true
	}

	if len(vp.splicePTS) == 0 {
		return true
	}

	spliceTime := vp.splicePTS[0]

	if uint64(data.Pts) > spliceTime {
		vp.logger.Error("No I frame at splice PTS %d", spliceTime)
		vp.splicePTS = vp.splicePTS[1:]
		return false
	} else if uint64(data.Pts) == spliceTime {
		vp.logger.Info("I frame found for splice PTS %d", spliceTime)
		vp.splicePTS = vp.splicePTS[1:]
		return true
	}
	return true
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
