package ioUtils

import (
	"errors"
	"fmt"
	"strings"
)

type _OUTPUT_TYPE int

const (
	_OUTPUT_FILE    _OUTPUT_TYPE = 1
	_OUTPUT_ADSMART _OUTPUT_TYPE = 2
)

type ioWriterParam struct {
	OutputType _OUTPUT_TYPE
	FileOutput fileOutputParam
}

type fileOutputParam struct {
	OutFolder        string // Folder to store output
	RawByteExtension string // Extension for raw_id in data buffer
}

func (ot *_OUTPUT_TYPE) UnmarshalJSON(b []byte) error {
	str := strings.Trim(string(b), `"`)

	switch {
	case str == "_OUTPUT_FILE":
		*ot = _OUTPUT_FILE
	case str == "_OUTPUT_ADSMART":
		*ot = _OUTPUT_ADSMART
	default:
		return errors.New(fmt.Sprintf("Unknown option %s", str))
	}

	return nil
}
