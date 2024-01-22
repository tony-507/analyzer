package protocol

import (
	"fmt"
	"strings"
)

type IParser interface {
	Parse(*ParseResult) []ParseResult // Parse given data
}

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

type ParseResult struct {
	Buffer  []byte
	Fields  map[string]int64
	IsEmpty bool
}

func (res *ParseResult) GetBuffer() []byte {
	return res.Buffer
}

func (res *ParseResult) GetField(name string) (int64, bool) {
	val, ok := res.Fields[name]
	return val, ok
}

func EmptyResult() ParseResult {
	return ParseResult{
		IsEmpty: true,
	}
}

func AssertIntEqual(name string, expected int, actual int) {
	if expected != actual {
		panic(fmt.Sprintf("Invalid value at %s. Expecting %d but got %d", name, expected, actual))
	}
}
