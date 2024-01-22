package model

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/tony-507/analyzers/src/tttKernel"
	"github.com/tony-507/analyzers/src/common/io"
)

type scte35Struct struct {
	callback   PsiManager
	header     tttKernel.CmBuf
	payload    []byte
	pid        int
	schema     *Scte35Schema
	sectionLen int
}

type Scte35Schema struct {
	PktCnt           int
	PtsAdjustment    int
	CwIdx            int
	Tier             int
	SpliceCmdLen     int
	SpliceCmdTypeStr string
	SpliceCmd        Splice_command
}

func (s *scte35Struct) Append(buf []byte) {
	s.payload = append(s.payload, buf...)
}

func (s *scte35Struct) GetField(str string) (int, error) {
	return resolveHeaderField(s, str)
}

func (s *scte35Struct) GetName() string {
	return "SCTE-35"
}

func (s *scte35Struct) GetHeader() tttKernel.CmBuf {
	return s.header
}

func (s *scte35Struct) GetPayload() []byte {
	return s.payload
}

func (s *scte35Struct) Ready() bool {
	return len(s.payload) >= s.sectionLen
}

func (s *scte35Struct) Serialize() []byte {
	// TODO
	return []byte{}
}

func (s *scte35Struct) setBuffer(inBuf []byte) error {
	buf := inBuf[:2]
	r := io.GetBufferReader(buf)
	s.header = tttKernel.MakeSimpleBuf(buf)
	if r.ReadBits(1) != 0 {
		return errors.New("Section syntax indicator of SCTE-35 splice info section is not set to 0")
	}
	if r.ReadBits(1) != 0 {
		return errors.New("Private bits of SCTE-35 splice info section is not set to 0")
	}
	r.ReadBits(2) // SAP type
	s.sectionLen = r.ReadBits(12)
	s.header.SetField("sectionLength", s.sectionLen, true)

	s.payload = inBuf[2:]
	return nil
}

func (s *scte35Struct) Process() error {
	schema := s.schema
	r := io.GetBufferReader(s.payload)

	if r.ReadBits(8) != 0 {
		return errors.New("SCTE-35 protocol version is not 0")
	}
	encryptedPkt := r.ReadBits(1) != 0
	if encryptedPkt {
		return errors.New("SCTE-35 splice info section is encrypted")
	}
	r.ReadBits(6) // Encryption algorithm
	schema.PtsAdjustment = r.ReadBits(33)
	schema.CwIdx = r.ReadBits(8)
	schema.Tier = r.ReadBits(12)

	schema.SpliceCmdLen = r.ReadBits(12)

	spliceCmdType := r.ReadBits(8)
	spliceCmdTypeStr := "Unknown"
	var spliceCmd Splice_command

	switch spliceCmdType {
	case 0x00:
		// Splice null, do nothing
		spliceCmdTypeStr = "splice_null"
		spliceCmd = Splice_null{}
	case 0x04:
		// Splice schedule
		spliceCmdTypeStr = "splice_schedule"
		spliceCmd = readSpliceSchedule(&r)
	case 0x05:
		// Splice insert
		spliceCmdTypeStr = "splice_insert"
		spliceCmd = readSpliceEvent(&r, true)
	case 0x06:
		// Time signal
		spliceCmdTypeStr = "time_signal"
		spliceCmd = readTimeSignal(&r)
	case 0x07:
		// Bandwidth reservation
		spliceCmdTypeStr = "bandwidth_reservation"
	case 0xff:
		// Private command
		spliceCmdTypeStr = "private_command"
		spliceCmd = readPrivateCommand(&r)
	default:
		msg := fmt.Sprintf("unknown splice command type %d received", spliceCmdType)
		return errors.New(msg)
	}

	s.callback.SpliceEventReceived(s.pid, spliceCmdTypeStr, spliceCmd.GetSplicePTS(), s.schema.PktCnt)

	schema.SpliceCmdTypeStr = spliceCmdTypeStr
	schema.SpliceCmd = spliceCmd

	s.schema = schema

	// TODO Continue...

	jsonBytes, _ := json.MarshalIndent(s.schema, "", "\t")
	s.callback.PsiUpdateFinished(s.pid, -1, jsonBytes)

	return nil
}

func Scte35Table(manager PsiManager, pktCnt int, pid int, buf []byte) (DataStruct, error) {
	rv := &scte35Struct{callback: manager, payload: make([]byte, 0), pid: pid}
	rv.schema = &Scte35Schema{PktCnt: pktCnt, SpliceCmd: nil}
	err := rv.setBuffer(buf)
	return rv, err
}
