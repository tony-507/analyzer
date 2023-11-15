package h264

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tony-507/analyzers/src/common/io"
	"github.com/tony-507/analyzers/src/plugins/dataHandler/utils"
	commonUtils "github.com/tony-507/analyzers/src/utils"
)

func TestReadPicTiming(t *testing.T) {
	rbsp := []byte{0x1b, 0x00, 0x02, 0xb9, 0x16, 0x00, 0x00, 0x00, 0x08}
	r := io.GetBufferReader(rbsp)
	sequenceParameterSet := CreateSequenceParameterSet()
	sequenceParameterSet.Vui.PicStructPresentFlag = true
	expectedPicTiming := PicTiming{
		PicStructPresentFlag: true,
		Clocks: []PicClock{
			{
				CtType: 1,
				CountingType: 0,
				DiscontinuityFlag: false,
				CntDroppedFlag: false,
				Tc: commonUtils.TimeCode{
					Hour: 0,
					Minute: 5,
					Second: 28,
					Frame: 2,
				},
			},
		},
	}
	picTiming := ParsePicTiming(&r, sequenceParameterSet)
	assert.Equal(t, expectedPicTiming, picTiming)
	assert.Equal(t, r.GetOffset(), 4)
	assert.Equal(t, r.GetPos(), 8)
}

func TestReadSliceHeader(t *testing.T) {
	rbsp := []byte{0x9e, 0x0e, 0x9f}
	r := io.GetBufferReader(rbsp)
	data := utils.CreateParsedData()
	vData := data.GetVideoData()
	readSliceHeader(&r, vData)

	assert.Equal(t, utils.B_SLICE, vData.Type)
}
