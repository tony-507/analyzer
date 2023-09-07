package def

import "strings"

const (
	TS_PKT_SIZE int = 188
)

type PROTOCOL int

const (
	PROT_UNKNOWN PROTOCOL = 0
	PROT_TS      PROTOCOL = 1
	PROT_RTP     PROTOCOL = 2
)

func StringToProtocol(prot_name string) PROTOCOL {
	switch strings.ToLower(prot_name) {
	case "ts":
		return PROT_TS
	case "rtp":
		return PROT_RTP
	default:
		return PROT_UNKNOWN
	}
}