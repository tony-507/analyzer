package model

import (
	"errors"
	"fmt"

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

	afc := r.ReadBits(2)

	p.header.SetField("afc", afc, true)
	p.header.SetField("cc", r.ReadBits(4), true)

	afSize := 0

	if afc > 1 {
		p.hasAdaptationField = true
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
	r := common.GetBufferReader(inBuf)

	afLen := r.ReadBits(8)
	if afLen == 0 {
		return 1, nil
	}

	p.adaptationField = common.MakeSimpleBuf(inBuf[:afLen])

	remainedLen := afLen
	p.adaptationField.SetField("discontinuityIdr", r.ReadBits(1), true)
	p.adaptationField.SetField("randomAccessIdr", r.ReadBits(1), true)
	p.adaptationField.SetField("esPriorityIdr", r.ReadBits(1), true)

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

	p.adaptationField.SetField("pcr", pcr, true)
	p.adaptationField.SetField("opcr", opcr, true)
	p.adaptationField.SetField("spliceCountdown", spliceCountdown, true)
	p.adaptationField.SetField("privateData", privateData, true)

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

func (p *tsPacketStruct) GetField(str string) (int, error) {
	return resolveHeaderField(p, str)
}

func (p *tsPacketStruct) GetHeader() common.CmBuf {
	return p.header
}

func (p *tsPacketStruct) GetValueFromAdaptationField(name string) (int, error) {
	field, ok := p.adaptationField.GetField(name)
	if !ok {
		return 0, errors.New(fmt.Sprintf("%s not exist in adaptation field", name))
	}
	val, isInt := field.(int)
	if !isInt {
		return 0, errors.New(fmt.Sprintf("%s is not an integer", name))
	}
	return val, nil
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

func TsPacket(buf []byte) (*tsPacketStruct, error) {
	rv := &tsPacketStruct{hasAdaptationField: false, payload: make([]byte, 0)}
	err := rv.setBuffer(buf)
	return rv, err
}
