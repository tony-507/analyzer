package model

import (
	"errors"

	"github.com/tony-507/analyzers/src/common/io"
)

type tsPacketHeader struct {
	Tei      bool
	Pusi     bool
	Priority bool
	Pid      int
	Tsc      int
	Afc      int
	Cc       int
}

type adaptationField struct {
	exist           bool
	Discontinuity   bool
	RandomAccess    bool
	EsPriority      bool
	Pcr             int64
	Opcr            int64
	SpliceCountdown int
	PrivateData     string
}

type tsPacketStruct struct {
	header             tsPacketHeader
	adaptationField    adaptationField
	payload            []byte
}

func (p *tsPacketStruct) setBuffer(inBuf []byte) error {
	buf := inBuf[:4]
	r := io.GetBufferReader(buf)
	if r.ReadBits(8) != 0x47 {
		return errors.New("TS sync byte not match")
	}
	p.header.Tei = r.ReadBits(1) != 0
	p.header.Pusi = r.ReadBits(1) != 0
	p.header.Priority = r.ReadBits(1) != 0
	p.header.Pid = r.ReadBits(13)
	p.header.Tsc = r.ReadBits(2)
	p.header.Afc = r.ReadBits(2)
	p.header.Cc = r.ReadBits(4)

	afSize := 0

	if p.header.Afc > 1 {
		p.adaptationField.exist = true
		var err error
		afSize, err = p.readAdaptationField(inBuf[4:])
		if err != nil {
			return err
		}
	}

	p.payload = inBuf[(4 + afSize):]

	return nil
}

func (p *tsPacketStruct) readAdaptationField(inBuf []byte) (int, error) {
	r := io.GetBufferReader(inBuf)

	afLen := r.ReadBits(8)
	if afLen == 0 {
		return 1, nil
	}

	remainedLen := afLen
	p.adaptationField.Discontinuity = r.ReadBits(1) != 0
	p.adaptationField.RandomAccess = r.ReadBits(1) != 0
	p.adaptationField.EsPriority = r.ReadBits(1) != 0

	pcrFlag := r.ReadBits(1)
	opcrFlag := r.ReadBits(1)
	spliceCountdownFlag := r.ReadBits(1)
	transportPrivateFlag := r.ReadBits(1)
	afExtensionFlag := r.ReadBits(1)
	remainedLen -= 1

	pcr := -1
	opcr := -1
	spliceCountdown := -1
	privateData := ""

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
		privateData = r.ReadChar(privateDataLen)
		remainedLen -= privateDataLen + 1
	}

	p.adaptationField.Pcr = int64(pcr)
	p.adaptationField.Opcr = int64(opcr)
	p.adaptationField.SpliceCountdown = spliceCountdown
	p.adaptationField.PrivateData = privateData

	if afExtensionFlag != 0 {
	}

	r.ReadBits(8 * remainedLen)

	return afLen + 1, nil
}

func (p *tsPacketStruct) Process() error {
	return nil
}

func (p *tsPacketStruct) Append(buf []byte) {
	p.payload = append(p.payload, buf...)
}

func (p *tsPacketStruct) GetHeader() tsPacketHeader {
	return p.header
}

func (p *tsPacketStruct) GetAdaptationField() adaptationField {
	return p.adaptationField
}

func (p *tsPacketStruct) GetPayload() []byte {
	return p.payload
}

func (p *tsPacketStruct) GetName() string {
	return "TS packet"
}

func (p *tsPacketStruct) Ready() bool {
	return len(p.payload) == 184
}

func (p *tsPacketStruct) HasAdaptationField() bool {
	return p.adaptationField.exist
}

func TsPacket(buf []byte) (*tsPacketStruct, error) {
	rv := &tsPacketStruct{
		header: tsPacketHeader{},
		adaptationField: adaptationField{ exist: false },
		payload: make([]byte, 0),
	}
	err := rv.setBuffer(buf)
	return rv, err
}
