package ioUtils

import (
	"errors"
	"fmt"
	"strings"
)

type _SOURCE_TYPE int

const (
	_SOURCE_DUMMY _SOURCE_TYPE = 0
	_SOURCE_FILE  _SOURCE_TYPE = 1
	_SOURCE_UDP   _SOURCE_TYPE = 2
)

type ioReaderParam struct {
	Source       _SOURCE_TYPE
	FileInput    fileInputParam
	UdpInput     udpInputParam
	SkipCnt      int  // Number of packets to skip at start
	MaxInCnt     int  // Number of packets to be parsed
	DumpRawInput bool // Dump input data
}

type fileInputParam struct {
	Fname string
}

type udpInputParam struct {
	Address string
	Itf     string
	Timeout int
}

func (st *_SOURCE_TYPE) UnmarshalJSON(b []byte) error {
	str := strings.Trim(string(b), `"`)

	switch {
	case str == "_SOURCE_DUMMY":
		*st = _SOURCE_DUMMY
	case str == "_SOURCE_FILE":
		*st = _SOURCE_FILE
	case str == "_SOURCE_UDP":
		*st = _SOURCE_UDP
	default:
		return errors.New(fmt.Sprintf("Unknown option %s", str))
	}

	return nil
}
