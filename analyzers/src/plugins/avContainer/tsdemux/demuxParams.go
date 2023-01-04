package tsdemux

import (
	"errors"
	"fmt"
	"strings"
)

type _DEMUX_MODE int

const (
	_DEMUX_DUMMY _DEMUX_MODE = 0
	_DEMUX_FULL  _DEMUX_MODE = 1
	_DEMUX_PSI   _DEMUX_MODE = 2
)

type demuxParams struct {
	Mode _DEMUX_MODE
}

func (dm *_DEMUX_MODE) UnmarshalJSON(b []byte) error {
	str := strings.Trim(string(b), `"`)

	switch {
	case str == "_DEMUX_DUMMY":
		*dm = _DEMUX_DUMMY
	case str == "_DEMUX_FULL":
		*dm = _DEMUX_FULL
	case str == "_DEMUX_PSI":
		*dm = _DEMUX_PSI
	default:
		return errors.New(fmt.Sprintf("Unknown option %s", str))
	}

	return nil
}
