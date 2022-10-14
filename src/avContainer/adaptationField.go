package avContainer

import (
	"fmt"

	"github.com/tony-507/analyzers/src/common"
)

type AdaptationField struct {
	AfLen        int
	Pcr          int
	Splice_point int
	Private_data string
}

func ParseAdaptationField(buf []byte) AdaptationField {
	actualAfLen := 0
	pcr := -1
	splice_point := -1
	private_data := ""

	r := common.GetBufferReader(buf)
	afLen := r.ReadBits(8)
	if afLen == 0 {
		return AdaptationField{Pcr: -1}
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

	return AdaptationField{AfLen: actualAfLen, Pcr: pcr, Splice_point: splice_point, Private_data: private_data}
}

// Adaptation field construction
func ConstructAdapationField(random_access bool, pcr int, opcr int, splice_point int, privateBytes []byte, stuffing []byte) []byte {
	// We need section length before actual construction
	actualAfLen := 2
	pcrFlag := 0
	opcrFlag := 0
	spliceFlag := 0
	privateFlag := 0
	stuffFlag := 0
	if pcr != -1 {
		actualAfLen += 6
		pcrFlag = 1
	}
	if opcr != -1 {
		actualAfLen += 6
		opcrFlag = 1
	}
	if splice_point != -1 {
		actualAfLen += 1
		spliceFlag = 1
	}
	if len(privateBytes) != 0 {
		actualAfLen += len(privateBytes)
		privateFlag = 1
	}
	if len(stuffing) != 0 {
		actualAfLen += len(stuffing)
		stuffFlag = 1
	}

	w := common.GetBufferWriter(actualAfLen)
	w.WriteInt(actualAfLen)
	w.Write(0, 1)

	if random_access {
		w.Write(1, 1)
	} else {
		w.Write(0, 1)
	}

	w.Write(0, 1)
	w.Write(pcrFlag, 1)
	w.Write(opcrFlag, 1)
	w.Write(spliceFlag, 1)
	w.Write(privateFlag, 1)
	w.Write(0, 1)

	if pcrFlag == 1 {
		w.Write(pcr/300, 33)
		w.Write(0x3f, 6)
		w.Write(pcr%300, 9)
	}
	if opcrFlag == 1 {
		w.Write(opcr/300, 33)
		w.Write(0x3f, 6)
		w.Write(opcr%300, 9)
	}
	if spliceFlag == 1 {
		w.Write(splice_point, 8)
	}
	if privateFlag == 1 {
		for _, data := range privateBytes {
			w.WriteByte(int(data))
		}
	}
	if stuffFlag == 1 {
		for i := 0; i < len(stuffing); i++ {
			w.WriteByte(0xff)
		}
	}

	return w.GetBuf()
}
