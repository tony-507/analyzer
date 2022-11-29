package tsdemux

type DEMUX_MODE int

const (
	DEMUX_DUMMY DEMUX_MODE = 0
	DEMUX_FULL  DEMUX_MODE = 1
	DEMUX_PSI   DEMUX_MODE = 2
)

type DemuxParams struct {
	Mode DEMUX_MODE
}
