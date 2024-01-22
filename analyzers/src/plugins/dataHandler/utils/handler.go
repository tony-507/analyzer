package utils

import (
	"strconv"

	"github.com/tony-507/analyzers/src/plugins/common"
	"github.com/tony-507/analyzers/src/tttKernel"
)

type PARSED_TYPE int

const (
	EMPTY        PARSED_TYPE = 0
	PARSED_VIDEO PARSED_TYPE = 1
	PARSED_AUDIO PARSED_TYPE = 2
	PARSED_DATA  PARSED_TYPE = 3
)

type ParsedData struct {
	dType  PARSED_TYPE
	vData VideoDataStruct
}

func (data *ParsedData) GetType() PARSED_TYPE {
	return data.dType
}

func (data *ParsedData) GetVideoData() *VideoDataStruct {
	data.dType = PARSED_VIDEO
	return &data.vData
}

func CreateParsedData() ParsedData {
	return ParsedData{
		dType: EMPTY,
		vData: VideoData(),
	}
}

type FRAME_TYPE int

const (
	UNKNOWN_SLICE FRAME_TYPE = 0
	I_SLICE       FRAME_TYPE = 1
	P_SLICE       FRAME_TYPE = 2
	B_SLICE       FRAME_TYPE = 3
)

type VideoDataStruct struct {
	Dts      int
	Pts      int
	TimeCode common.TimeCode
	Type     FRAME_TYPE
}

func (d * VideoDataStruct) GetType() PARSED_TYPE {
	return PARSED_VIDEO
}

func (d *VideoDataStruct) ToCmBuf() tttKernel.CmBuf {
	cmBuf := tttKernel.MakeSimpleBuf([]byte{})
	if d.Type != UNKNOWN_SLICE {
		cmBuf.SetField("type", strconv.Itoa(int(d.Type)), false)
	}
	cmBuf.SetField("pts", d.Pts, false)
	if !d.TimeCode.IsEmpty() {
		cmBuf.SetField("timecode", d.TimeCode.ToString(), false)
	}
	return cmBuf
}

func VideoData() VideoDataStruct {
	return VideoDataStruct{
		TimeCode: common.NewTimeCode(),
		Type: UNKNOWN_SLICE,
	}
}

type DataHandler interface {
	Feed(unit tttKernel.CmUnit, newData *ParsedData) error// Accept input buffer and begin parsing
}
