package model

import (
	"errors"

	"github.com/tony-507/analyzers/src/plugins/common/io"
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
	p.adaptationField = newAdaptationField()
	return p.adaptationField.read(inBuf)
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
