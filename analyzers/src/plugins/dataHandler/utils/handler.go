package utils

import (
	"github.com/tony-507/analyzers/src/common"
	"github.com/tony-507/analyzers/src/utils"
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
	UNKNOWN FRAME_TYPE = 0
	I       FRAME_TYPE = 1
	P       FRAME_TYPE = 2
	B       FRAME_TYPE = 3
)

type VideoDataStruct struct {
	Dts      int
	Pts      int
	TimeCode utils.TimeCode
	Type     FRAME_TYPE
}

func (d * VideoDataStruct) GetType() PARSED_TYPE {
	return PARSED_VIDEO
}

func VideoData() VideoDataStruct {
	return VideoDataStruct{
		TimeCode: utils.TimeCode{Frame: -1},
		Type: UNKNOWN,
	}
}

type DataHandler interface {
	Feed(unit common.CmUnit, newData *ParsedData) error// Accept input buffer and begin parsing
}
