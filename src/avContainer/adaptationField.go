package avContainer

import (
	"fmt"

	"github.com/tony-507/analyzers/src/common"
)

type AdaptationField struct {
	afLen        int
	pcr          int
	splice_point int
	private_data string
}

func ParseAdaptationField(buf []byte) AdaptationField {
	actualAfLen := 0
	pcr := -1
	splice_point := -1
	private_data := ""

	r := common.GetBufferReader(buf)
	afLen := r.ReadBits(8)
	if afLen == 0 {
		return AdaptationField{pcr: -1}
	} else {
		actualAfLen = afLen + 1
	}
	r.ReadBits(1) // Discontinuity counter
	r.ReadBits(1) // Random access indicator
	r.ReadBits(1) // ES indicator
	pcr_flag := r.ReadBits(1) != 0
	opcr_flag := r.ReadBits(1) != 0
	splice_flag := r.ReadBits(1) != 0
	private_flag := r.ReadBits(1) != 0
	r.ReadBits(1) // Adaptation field extension flag
	afLen -= 1

	if pcr_flag {
		pcr = r.ReadBits(33)
		r.ReadBits(6)                 // Reserved
		pcr = pcr*300 + r.ReadBits(9) // Extension
		afLen -= 6
	}
	if opcr_flag {
		r.ReadBits(48)
		afLen -= 6
	}
	if splice_flag {
		splice_point = r.ReadBits(8)
		afLen -= 1
	}
	if private_flag {
		dataLen := r.ReadBits(8)
		for i := 0; i < dataLen; i++ {
			private_data += fmt.Sprint(r.ReadBits(8))
		}
	}

	r.ReadBits(afLen)

	return AdaptationField{afLen: actualAfLen, pcr: pcr, splice_point: splice_point, private_data: private_data}
}
