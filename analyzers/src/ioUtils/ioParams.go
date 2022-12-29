package ioUtils

import (
	"errors"
	"strings"
)

type SOURCE_TYPE int
type OUTPUT_TYPE int

const (
	SOURCE_DUMMY SOURCE_TYPE = 0
	SOURCE_FILE  SOURCE_TYPE = 1
)

const (
	OUTPUT_FILE OUTPUT_TYPE = 1
)

type IOReaderParam struct {
	Source    SOURCE_TYPE
	FileInput FileInputParam
	SkipCnt   int // Number of packets to skip at start
	MaxInCnt  int // Number of packets to be parsed
}

type FileInputParam struct {
	Fname string
}

func (st *SOURCE_TYPE) UnmarshalJSON(b []byte) error {
	str := strings.Trim(string(b), `"`)

	switch {
	case str == "SOURCE_DUMMY":
		*st = SOURCE_DUMMY
	case str == "SOURCE_FILE":
		*st = SOURCE_FILE
	default:
		return errors.New("Unknown option")
	}

	return nil
}

type IOWriterParam struct {
	OutputType OUTPUT_TYPE
	FileOutput FileOutputParam
}

type FileOutputParam struct {
	OutFolder string // Folder to store output
}

func (ot *OUTPUT_TYPE) UnmarshalJSON(b []byte) error {
	str := strings.Trim(string(b), `"`)

	switch {
	case str == "OUTPUT_FILE":
		*ot = OUTPUT_FILE
	default:
		return errors.New("Unknown option")
	}

	return nil
}
