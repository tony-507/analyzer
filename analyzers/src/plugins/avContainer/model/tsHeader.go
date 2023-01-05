package model

import (
	"github.com/tony-507/analyzers/src/common"
)

// Keep header data I want
type TsHeader struct {
	Tei      bool // Transport error indicator
	Pusi     bool // For message support
	Priority bool // Transport priority
	Pid      int
	Tsc      int // Scrambling control
	Afc      int // Adaptation field control
	Cc       int // Continuity counter
}

func ReadTsHeader(buf []byte) TsHeader {
	r := common.GetBufferReader(buf)
	if r.ReadBits(8) != 0x47 {
		panic("TS packet sync byte not match")
	}
	tei := r.ReadBits(1) != 0
	pusi := r.ReadBits(1) != 0
	priority := r.ReadBits(1) != 0
	pid := r.ReadBits(13)
	tsc := r.ReadBits(2)
	afc := r.ReadBits(2)
	cc := r.ReadBits(4)
	return TsHeader{Tei: tei, Pusi: pusi, Priority: priority, Pid: pid, Tsc: tsc, Afc: afc, Cc: cc}
}

// TsHeader construction
func (th *TsHeader) Serialize() []byte {
	w := common.GetBufferWriter(4)

	w.WriteByte(0x47)
	if th.Tei {
		w.Write(1, 1)
	} else {
		w.Write(0, 1)
	}
	if th.Pusi {
		w.Write(1, 1)
	} else {
		w.Write(0, 1)
	}
	if th.Priority {
		w.Write(1, 1)
	} else {
		w.Write(0, 1)
	}
	w.Write(th.Pid, 13)
	w.Write(th.Tsc, 2)
	w.Write(th.Afc, 2)
	w.Write(th.Cc, 4)

	return w.GetBuf()
}
