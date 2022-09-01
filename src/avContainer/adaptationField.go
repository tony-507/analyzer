package avContainer

import (
	"github.com/tony-507/analyzers/src/common"
)

type AdaptationField struct {
	pcr          int
	splice_point int
	private_data string
}

func ParseAdaptationField(r *common.BsReader) AdaptationField {
	pcr := -1
	splice_point := -1
	private_data := ""

	afLen := (*r).ReadBits(8)
	(*r).ReadBits(1) // Discontinuity counter
	(*r).ReadBits(1) // Random access indicator
	(*r).ReadBits(1) // ES indicator
	pcr_flag := (*r).ReadBits(1) != 0
	opcr_flag := (*r).ReadBits(1) != 0
	splice_flag := (*r).ReadBits(1) != 0
	private_flag := (*r).ReadBits(1) != 0
	(*r).ReadBits(1) // Adaptation field extension flag
	afLen -= 1

	if pcr_flag {
		pcr = (*r).ReadBits(33)
		(*r).ReadBits(6)                 // Reserved
		pcr = pcr*300 + (*r).ReadBits(9) // Extension
		afLen -= 6
	}
	if opcr_flag {
		(*r).ReadBits(48)
		afLen -= 6
	}
	if splice_flag {
		splice_point = (*r).ReadBits(8)
		afLen -= 1
	}
	if private_flag {
		dataLen := (*r).ReadBits(8)
		afLen -= dataLen + 1
		for i := 0; i < dataLen; i++ {
			private_data += string((*r).ReadBits(8))
		}
	}

	(*r).ReadBits(afLen)

	return AdaptationField{pcr, splice_point, private_data}
}
