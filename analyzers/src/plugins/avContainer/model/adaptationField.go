package model

import (
	"github.com/tony-507/analyzers/src/plugins/common/io"
)

type adaptationField struct {
	exist           bool
	Discontinuity   bool
	RandomAccess    bool
	EsPriority      bool
	Pcr             int64
	Opcr            int64
	SpliceCountdown int
	PrivateData     []privateData
}

func (af *adaptationField) read(buf []byte) (int, error) {
	r := io.GetBufferReader(buf)

	afLen := r.ReadBits(8)
	if afLen == 0 {
		return 1, nil
	}

	remainedLen := afLen
	af.Discontinuity = r.ReadBits(1) != 0
	af.RandomAccess = r.ReadBits(1) != 0
	af.EsPriority = r.ReadBits(1) != 0

	pcrFlag := r.ReadBits(1)
	opcrFlag := r.ReadBits(1)
	spliceCountdownFlag := r.ReadBits(1)
	transportPrivateFlag := r.ReadBits(1)
	afExtensionFlag := r.ReadBits(1)
	remainedLen -= 1

	pcr := -1
	opcr := -1
	spliceCountdown := -1
	privateData := []privateData{}

	if pcrFlag != 0 {
		pcr = r.ReadBits(33)
		r.ReadBits(6)                 // Reserved
		pcr = pcr*300 + r.ReadBits(9) // Extension
		remainedLen -= 6
	}

	if opcrFlag != 0 {
		opcr = r.ReadBits(33)
		r.ReadBits(6)                   // Reserved
		opcr = opcr*300 + r.ReadBits(9) // Extension
		remainedLen -= 6
	}

	if spliceCountdownFlag != 0 {
		spliceCountdown = r.ReadBits(8)
		remainedLen -= 1
	}

	if transportPrivateFlag != 0 {
		privateDataLen := r.ReadBits(8)
		remainedLen -= privateDataLen + 1
		for privateDataLen > 0 {
			pd, err := parsePrivateData(&r)
			if err != nil {
				panic(err)
			}
			privateData = append(privateData, pd)
			privateDataLen -= pd.Length + 2
		}
	}

	af.Pcr = int64(pcr)
	af.Opcr = int64(opcr)
	af.SpliceCountdown = spliceCountdown
	af.PrivateData = privateData

	if afExtensionFlag != 0 {
	}

	r.ReadBits(8 * remainedLen)

	return afLen + 1, nil
}

func newAdaptationField() adaptationField {
	return adaptationField{
		exist: true,
	}
}
