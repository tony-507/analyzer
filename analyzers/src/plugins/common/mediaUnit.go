package common

import (
	"fmt"

	"github.com/tony-507/analyzers/src/tttKernel"
)

// Units representing media data

// Video
type videoMetaData struct {
	Tc        TimeCode
	Type      FRAME_TYPE
	FrameRate _FRAME_RATE
	PicFlag   _PIC_FLAG
}

func NewVideoData() *videoMetaData {
	return &videoMetaData{
		Tc: NewTimeCode(),
		Type: UNKNOWN_SLICE,
	}
}

type TimeCode struct {
	Hour      int
	Minute    int
	Second    int
	Frame     int
	DropFrame bool
	Field     bool
}

func (tc *TimeCode) ToString() string {
	return fmt.Sprintf("%02d:%02d:%02d:%02d", tc.Hour, tc.Minute, tc.Second, tc.Frame)
}

func (tc *TimeCode) Equals(other *TimeCode) bool {
	return tc.Hour == other.Hour && tc.Minute == other.Minute && tc.Second == other.Second && tc.Frame == other.Frame
}

func (tc *TimeCode) IsEmpty() bool {
	emptyTc := NewTimeCode()
	return tc.Equals(&emptyTc)
}

func NewTimeCode() TimeCode {
	return TimeCode{-1, -1, -1, -1, false, false}
}

type FRAME_TYPE int

const (
	UNKNOWN_SLICE FRAME_TYPE = 0
	I_SLICE       FRAME_TYPE = 1
	P_SLICE       FRAME_TYPE = 2
	B_SLICE       FRAME_TYPE = 3
)

type _FRAME_RATE struct {
	num int
	den int
}

func FrameRate(num int, den int) _FRAME_RATE {
	return _FRAME_RATE{
		num: num,
		den: den,
	}
}

type _PIC_FLAG int

const (
	FRAME        _PIC_FLAG = 0
	TOP_FIELD    _PIC_FLAG = 1
	BOTTOM_FIELD _PIC_FLAG = 2
)

// Data structs
type mediaData interface {
	GetType() _DATA_TYPE
}

type _DATA_TYPE int
const (
	SCTE_35 _DATA_TYPE = 0
)

type scte35Data struct {
	SpliceTime int64
	Preroll    int64
}

func (data *scte35Data) GetType() _DATA_TYPE {
	return SCTE_35
}

func NewScte35Data(spliceTime int64, preroll int64) scte35Data {
	return scte35Data{
		SpliceTime: spliceTime,
		Preroll: preroll,
	}
}

// Buffer

type _MEDIA_TYPE int
const (
	UNKNOWN_UNIT _MEDIA_TYPE = 0
	VIDEO_UNIT   _MEDIA_TYPE = 1
	AUDIO_UNIT   _MEDIA_TYPE = 2
	DATA_UNIT    _MEDIA_TYPE = 3
)

type MediaUnit struct {
	buf      tttKernel.CmBuf
	unitType _MEDIA_TYPE
	vmd      *videoMetaData
	Data     mediaData
}

func NewMediaUnit(buf tttKernel.CmBuf, unitType _MEDIA_TYPE) *MediaUnit {
	return &MediaUnit{
		buf:      buf,
		unitType: unitType,
		vmd:      nil,
		Data:     nil,
	}
}

func (m *MediaUnit) GetBuf() tttKernel.CmBuf {
	return m.buf
}

func (m *MediaUnit) GetField(field string) interface{} {
	switch field {
	case "type":
		return m.unitType
	default:
		return nil
	}
}

func (m *MediaUnit) GetVideoData() *videoMetaData {
	if m.unitType == VIDEO_UNIT && m.vmd == nil {
		m.vmd = NewVideoData()
	}
	return m.vmd
}