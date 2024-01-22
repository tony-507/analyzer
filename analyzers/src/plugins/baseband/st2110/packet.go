package st2110

import "github.com/tony-507/analyzers/src/plugins/common/protocol"

type rtpPacket struct {
	payload     []byte
	payloadType int
	timestamp   uint32
	marker      bool
}

func newRtpPacket(rawBuffer []byte) rtpPacket {
	parser := protocol.GetParser(protocol.PROT_RTP)
	res := parser.Parse(&protocol.ParseResult{Buffer: rawBuffer})[0]

	pt, _ := res.GetField("payloadType")
	rtp, _ := res.GetField("timestamp")
	marker, _ := res.GetField("marker")

	return rtpPacket{
		payload: res.Buffer,
		payloadType: int(pt),
		timestamp: uint32(rtp),
		marker: marker != 0,
	}
}
