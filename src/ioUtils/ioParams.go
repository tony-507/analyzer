package ioUtils

type SOURCE_TYPE int

const (
	SOURCE_DUMMY SOURCE_TYPE = 0
	SOURCE_FILE  SOURCE_TYPE = 1
)

type IOReaderParam struct {
	Source SOURCE_TYPE
	FileInput FileInputParam
	SkipCnt int // Number of packets to skip at start
	MaxInCnt int // Number of packets to be parsed
}

type FileInputParam struct {
	Fname string
}

type IOWriterParam struct {
	OutFolder string // Folder to store output
}
