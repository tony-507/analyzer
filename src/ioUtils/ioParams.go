package ioUtils

type IOReaderParam struct {
	Fname string
	SkipCnt int // Number of packets to skip at start
	MaxInCnt int // Number of packets to be parsed
}

type IOWriterParam struct {
	OutFolder string // Folder to store output
}
