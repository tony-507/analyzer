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

func ReadTsHeader(r *common.BsReader) TsHeader {
	if (*r).ReadBits(8) != 0x47 {
		panic("TS packet sync byte not match")
	}
	tei := (*r).ReadBits(1) != 0
	pusi := (*r).ReadBits(1) != 0
	priority := (*r).ReadBits(1) != 0
	pid := (*r).ReadBits(13)
	tsc := (*r).ReadBits(2)
	afc := (*r).ReadBits(2)
	cc := (*r).ReadBits(4)
	return TsHeader{Tei: tei, Pusi: pusi, Priority: priority, Pid: pid, Tsc: tsc, Afc: afc, Cc: cc}
}

// TsHeader construction
func ConstructTsHeader(pusi int, pid int, afc int, cc int) []byte {
	w := common.GetBufferWriter(4)

	w.WriteByte(0x47)
	w.Write(0, 1)
	w.Write(pusi, 1)
	w.Write(0, 1)
	w.Write(pid, 13)
	w.Write(0, 2)
	w.Write(afc, 2)
	w.Write(cc, 4)

	return w.GetBuf()
}
