package h264

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tony-507/analyzers/src/common"
	"github.com/tony-507/analyzers/src/utils"
)

func TestReadPicTiming(t *testing.T) {
	rbsp := []byte{0x1b, 0x00, 0x02, 0xb9, 0x16, 0x00, 0x00, 0x00, 0x08}
	r := common.GetBufferReader(rbsp)
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
				Tc: utils.TimeCode{
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
