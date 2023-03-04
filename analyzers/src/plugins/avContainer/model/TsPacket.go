package model

import (
	"errors"

	"github.com/tony-507/analyzers/src/common"
)

type tsPacketStruct struct {
	header             common.CmBuf
	adaptationField    common.CmBuf
	payload            []byte
	hasAdaptationField bool
}

func (p *tsPacketStruct) setBuffer(inBuf []byte) error {
	buf := inBuf[:4]
	r := common.GetBufferReader(buf)
	if r.ReadBits(8) != 0x47 {
		return errors.New("TS sync byte not match")
	}
	p.header = common.MakeSimpleBuf(buf)
	p.header.SetField("tei", r.ReadBits(1), true)
	p.header.SetField("pusi", r.ReadBits(1), true)
	p.header.SetField("priority", r.ReadBits(1), true)
	p.header.SetField("pid", r.ReadBits(13), true)
	p.header.SetField("tsc", r.ReadBits(2), true)
	p.header.SetField("afc", r.ReadBits(2), true)
	p.header.SetField("cc", r.ReadBits(4), true)

	p.payload = inBuf[4:]

	return nil
}

func (p *tsPacketStruct) Append(buf []byte) {
	p.payload = append(p.payload, buf...)
}

func (p *tsPacketStruct) GetField(str string) (int, error) {
	return resolveHeaderField(p, str)
}

func (p *tsPacketStruct) getHeader() common.CmBuf {
	return p.header
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

func (p *tsPacketStruct) Serialize() []byte {
	// TODO
	return []byte{}
}

func (p *tsPacketStruct) HasAdaptationField() bool {
	return p.hasAdaptationField
}

func TsPacket(buf []byte) (DataStruct, error) {
	rv := &tsPacketStruct{hasAdaptationField: false, payload: make([]byte, 0)}
	err := rv.setBuffer(buf)
	return rv, err
}