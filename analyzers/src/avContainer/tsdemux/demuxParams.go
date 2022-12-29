package tsdemux

import (
	"errors"
	"strings"
)

type DEMUX_MODE int

const (
	DEMUX_DUMMY DEMUX_MODE = 0
	DEMUX_FULL  DEMUX_MODE = 1
	DEMUX_PSI   DEMUX_MODE = 2
)

type DemuxParams struct {
	Mode DEMUX_MODE
}

func (dm *DEMUX_MODE) UnmarshalJSON(b []byte) error {
	str := strings.Trim(string(b), `"`)

	switch {
	case str == "DEMUX_DUMMY":
		*dm = DEMUX_DUMMY
	case str == "DEMUX_FULL":
		*dm = DEMUX_FULL
	case str == "DEMUX_PSI":
		*dm = DEMUX_PSI
	default:
		return errors.New("Unknown option")
	}

	return nil
}
