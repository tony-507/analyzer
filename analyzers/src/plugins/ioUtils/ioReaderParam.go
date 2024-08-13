package ioUtils

type ioReaderParam struct {
	Uri          string
	SkipCnt      int    // Number of packets to skip at start
	MaxInCnt     int    // Number of packets to be parsed
	DumpRawInput bool   // Dump input data
	Protocols    string // List of application protocols used, e.g. TS over RTP over SRT would be SRT,RTP,TS
}

type fileInputParam struct {
	Fname string
}

type udpInputParam struct {
	Address string
	Itf     string
	Timeout int
}
