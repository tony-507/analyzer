package model

import (
	"github.com/tony-507/analyzers/src/common"
)

type AdaptationField struct {
	AfLen        int
	DisCnt_cnt   int
	RandomAccess int
	EsIdr        int
	Pcr          int
	Opcr         int
	Splice_point int
	Private_data string
	StuffSize    int
}

func ParseAdaptationField(buf []byte) AdaptationField {
	actualAfLen := 0
	pcr := -1
	opcr := -1
	splice_point := -1
	private_data := ""

	r := common.GetBufferReader(buf)
	afLen := r.ReadBits(8)
	if afLen == 0 {
		return AdaptationField{AfLen: 0}
	} else {
		actualAfLen = afLen
	}
	disCnt_cnt := r.ReadBits(1)   // Discontinuity counter
	randomAccess := r.ReadBits(1) // Random access indicator
	esIdr := r.ReadBits(1)        // ES indicator
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
		opcr = r.ReadBits(33)
		r.ReadBits(6)                   // Reserved
		opcr = opcr*300 + r.ReadBits(9) // Extension
		afLen -= 6
	}
	if splice_flag {
		splice_point = r.ReadBits(8)
		afLen -= 1
	}
	if private_flag {
		dataLen := r.ReadBits(8)
		for i := 0; i < dataLen; i++ {
			private_data += string(rune(r.ReadBits(8)))
		}
		afLen -= dataLen + 1
	}

	r.ReadBits(afLen)

	return AdaptationField{AfLen: actualAfLen, DisCnt_cnt: disCnt_cnt, RandomAccess: randomAccess, EsIdr: esIdr, Pcr: pcr, Opcr: opcr, Splice_point: splice_point, Private_data: private_data, StuffSize: afLen}
}

// Adaptation field construction
func (af *AdaptationField) Serialize() []byte {
	if af.AfLen == 0 {
		return []byte{0x00}
	}
	// We need section length before actual construction
	actualAfLen := 2
	pcrFlag := 0
	opcrFlag := 0
	spliceFlag := 0
	privateFlag := 0
	stuffFlag := 0
	if af.Pcr != -1 {
		actualAfLen += 6
		pcrFlag = 1
	}
	if af.Opcr != -1 {
		actualAfLen += 6
		opcrFlag = 1
	}
	if af.Splice_point != -1 {
		actualAfLen += 1
		spliceFlag = 1
	}
	if len(af.Private_data) != 0 {
		actualAfLen += len(af.Private_data) + 1
		privateFlag = 1
	}
	if af.StuffSize != 0 {
		actualAfLen += af.StuffSize
		stuffFlag = 1
	}

	w := common.GetBufferWriter(actualAfLen)
	w.WriteByte(actualAfLen - 1) // Omit this byte
	w.Write(0, 1)

	w.Write(af.RandomAccess, 1)

	w.Write(0, 1)
	w.Write(pcrFlag, 1)
	w.Write(opcrFlag, 1)
	w.Write(spliceFlag, 1)
	w.Write(privateFlag, 1)
	w.Write(0, 1)

	if pcrFlag == 1 {
		w.Write(af.Pcr/300, 33)
		w.Write(0x3f, 6)
		w.Write(af.Pcr%300, 9)
	}
	if opcrFlag == 1 {
		w.Write(af.Opcr/300, 33)
		w.Write(0x3f, 6)
		w.Write(af.Opcr%300, 9)
	}
	if spliceFlag == 1 {
		w.Write(af.Splice_point, 8)
	}
	if privateFlag == 1 {
		w.WriteByte(len(af.Private_data))
		for _, data := range af.Private_data {
			w.WriteByte(int(data))
		}
	}
	if stuffFlag == 1 {
		for i := 0; i < af.StuffSize; i++ {
			w.WriteByte(0xff)
		}
	}

	return w.GetBuf()
}
