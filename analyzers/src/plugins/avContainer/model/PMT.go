package model

import (
	"encoding/json"
	"errors"

	"github.com/tony-507/analyzers/src/common"
)

type Descriptor struct {
	Tag     int
	Content string
}

type DataStream struct {
	StreamPid  int
	StreamType int
	StreamDesc []Descriptor
}

type PmtSchema struct {
	PktCnt   int
	Version  int
	ProgDesc []Descriptor
	Streams  []DataStream
	Crc32    int
}

type PmtStruct struct {
	callback   PsiManager
	header     common.CmBuf
	payload    []byte
	schema     *PmtSchema
	sectionLen int
}

func (p *PmtStruct) Append(buf []byte) {
	p.payload = append(p.payload, buf...)
}

func (p *PmtStruct) GetField(str string) (int, error) {
	return resolveHeaderField(p, str)
}

func (p *PmtStruct) GetName() string {
	return "PMT"
}

func (p *PmtStruct) GetHeader() common.CmBuf {
	return p.header
}

func (p *PmtStruct) GetPayload() []byte {
	return p.payload
}

func (p *PmtStruct) Ready() bool {
	return len(p.payload) >= p.sectionLen
}

func (p *PmtStruct) Serialize() []byte {
	// TODO
	return []byte{}
}

func (p *PmtStruct) setBuffer(inBuf []byte) error {
	buf := inBuf[:2]
	r := common.GetBufferReader(buf)
	p.header = common.MakeSimpleBuf(buf)
	if r.ReadBits(1) != 1 {
		return errors.New("Section syntax indicator of PMT is not set to 1")
	}
	if r.ReadBits(1) != 0 {
		return errors.New("Private bits of PMT is not set to 0")
	}
	if r.ReadBits(2) != 3 {
		return errors.New("Reserved bits of PMT is not set to all 1s")
	}
	if r.ReadBits(2) != 0 {
		return errors.New("Unused bits of PMT is not set to all 0s")
	}
	p.sectionLen = r.ReadBits(10)
	p.header.SetField("sectionLength", p.sectionLen, true)

	p.payload = inBuf[2:]
	return nil
}

func (p *PmtStruct) Process() error {
	r := common.GetBufferReader(p.payload)
	remainedLen := p.sectionLen

	progNum := r.ReadBits(16)
	if r.ReadBits(2) != 3 {
		return errors.New("Reserved bits of PMT is not set to all 1s")
	}
	p.schema.Version = r.ReadBits(5)
	if p.callback.GetPmtVersion(progNum) == p.schema.Version {
		return nil
	}

	r.ReadBits(1)  // Current/ next indicator
	r.ReadBits(16) // section number and last section number

	remainedLen -= 9
	if r.ReadBits(3) != 7 {
		return errors.New("Reserved bits of PMT is not set to all 1s")
	}
	r.ReadBits(13) // PCR pid
	if r.ReadBits(4) != 15 {
		return errors.New("Reserved bits of PMT is not set to all 1s")
	}
	r.ReadBits(2) // Program info unused bits
	programInfoLen := r.ReadBits(10)
	remainedLen -= 4 + programInfoLen

	programDescriptors := make([]Descriptor, 0)
	for {
		if programInfoLen <= 0 {
			break
		}
		desc := _readDescriptor(&r, &programInfoLen)
		programDescriptors = append(programDescriptors, desc)
	}
	p.schema.ProgDesc = programDescriptors

	streams := make([]DataStream, 0)
	for {
		if remainedLen == 0 {
			break
		}
		streamType := r.ReadBits(8)
		if r.ReadBits(3) != 7 {
			return errors.New("Reserved bits of PMT is not set to all 1s")
		}
		streamPid := r.ReadBits(13)
		if r.ReadBits(4) != 15 {
			return errors.New("Reserved bits of PMT is not set to all 1s")
		}
		if r.ReadBits(2) != 0 {
			return errors.New("Unused bits of PMT is not set to all 0s")
		}
		esInfoLen := r.ReadBits(10)
		remainedLen -= 5 + esInfoLen

		streamDescriptors := make([]Descriptor, 0)
		for {
			if esInfoLen <= 0 {
				break
			}
			desc := _readDescriptor(&r, &esInfoLen)
			streamDescriptors = append(streamDescriptors, desc)
		}
		p.callback.AddStream(p.schema.Version, progNum, streamPid, streamType)
		streams = append(streams, DataStream{StreamPid: streamPid, StreamType: streamType,
			StreamDesc: streamDescriptors})
	}
	p.schema.Streams = streams

	p.schema.Crc32 = r.ReadBits(32)

	pmtPid := p.callback.GetPmtPidByProgNum(progNum)
	jsonBytes, _ := json.MarshalIndent(p.schema, "\t", "\t")
	p.callback.PsiUpdateFinished(pmtPid, jsonBytes)

	return nil
}

func PmtTable(manager PsiManager, pktCnt int, buf []byte) (DataStruct, error) {
	rv := &PmtStruct{callback: manager, payload: make([]byte, 0), sectionLen: -1}
	rv.schema = &PmtSchema{PktCnt: pktCnt, Version: -1,
		ProgDesc: make([]Descriptor, 0), Streams: make([]DataStream, 0), Crc32: -1}
	err := rv.setBuffer(buf)
	return rv, err
}

func _readDescriptor(r *common.BsReader, l *int) Descriptor {
	Tag := (*r).ReadBits(8)
	descLen := (*r).ReadBits(8)
	Content := ""
	if descLen > 0 {
		Content = (*r).ReadHex(descLen)
	}
	*l -= descLen + 2
	return Descriptor{Tag, Content}
}
