package avContainer

import (
	"github.com/tony-507/analyzers/src/common"
)

// Keep header data I want
type TsHeader struct {
	tei      bool // Transport error indicator
	pusi     bool // For message support
	priority bool // Transport priority
	pid      int
	tsc      int // Scrambling control
	afc      int // Adaptation field control
	cc       int // Continuity counter
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
	return TsHeader{tei, pusi, priority, pid, tsc, afc, cc}
}
