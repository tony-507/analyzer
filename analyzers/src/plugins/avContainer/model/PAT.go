package model

// Parsing and writing of PAT
// Known issue:
// * Not support PAT with size > 1 TS packet

import (
	"encoding/json"
	"errors"

	"github.com/tony-507/analyzers/src/common/io"
	"github.com/tony-507/analyzers/src/tttKernel"
)

type patStruct struct {
	callback   PsiManager
	header     tttKernel.CmBuf
	payload    []byte
	schema     *PatSchema
	sectionLen int
}

type PatSchema struct {
	PktCnt     int
	Version    int
	ProgramMap map[int]int
	Crc32      int
}

func (p *patStruct) setBuffer(inBuf []byte) error {
	buf := inBuf[:2]
	r := io.GetBufferReader(buf)
	p.header = tttKernel.MakeSimpleBuf(buf)
	if r.ReadBits(1) != 1 {
		return errors.New("Section syntax indicator of PAT is not set to 1")
	}
	if r.ReadBits(1) != 0 {
		return errors.New("Private bits of PAT is not set to 0")
	}
	if r.ReadBits(2) != 3 {
		return errors.New("Reserved bits of PAT is not set to all 1s")
	}
	if r.ReadBits(2) != 0 {
		return errors.New("Unused bits of PAT is not set to all 0s")
	}
	p.sectionLen = r.ReadBits(10)
	p.header.SetField("sectionLength", p.sectionLen, true)

	p.payload = inBuf[2:]
	return nil
}

func (p *patStruct) Process() error {
	remainedLen := p.sectionLen
	r := io.GetBufferReader(p.payload)

	r.ReadBits(16) // Table Id extension
	if r.ReadBits(2) != 3 {
		return errors.New("Reserved bits of PAT is not set to all 1s")
	}
	p.schema.Version = r.ReadBits(5)
	if p.callback.GetPATVersion() == p.schema.Version {
		return nil
	}
	r.ReadBits(1)  // current/ next indicator
	r.ReadBits(16) // section number and last section number

	remainedLen -= 9
	for {
		if remainedLen <= 0 {
			break
		}
		progNum := r.ReadBits(16)
		if r.ReadBits(3) != 7 {
			return errors.New("Reserved bits of PAT is not set to all 1s")
		}
		pid := r.ReadBits(13)
		p.callback.AddProgram(p.schema.Version, progNum, pid)
		p.schema.ProgramMap[progNum] = pid
		remainedLen -= 4
	}
	if remainedLen < 0 {
		// Protection
		return errors.New("Something wrong with section length")
	}
	p.schema.Crc32 = r.ReadBits(4)

	jsonBytes, _ := json.MarshalIndent(p.schema, "", "\t") // Extra tab prefix to support array of Jsons

	p.callback.PsiUpdateFinished(0, p.schema.Version, jsonBytes)

	return nil
}

func (p *patStruct) Append(payload []byte) {
	p.payload = append(p.payload, payload...)
}

func (p *patStruct) GetField(str string) (int, error) {
	return resolveHeaderField(p, str)
}

func (p *patStruct) GetName() string {
	return "PAT"
}

func (p *patStruct) GetHeader() tttKernel.CmBuf {
	return p.header
}

func (p *patStruct) GetPayload() []byte {
	return p.payload
}

func (p *patStruct) Ready() bool {
	return len(p.payload) >= p.sectionLen
}

func (p *patStruct) Serialize() []byte {
	// TODO
	return []byte{}
}

func PatTable(manager PsiManager, pktCnt int, buf []byte) (DataStruct, error) {
	rv := &patStruct{callback: manager, payload: make([]byte, 0)}
	rv.schema = &PatSchema{PktCnt: pktCnt, Version: -1, ProgramMap: make(map[int]int, 0), Crc32: -1}
	err := rv.setBuffer(buf)
	return rv, err
}
