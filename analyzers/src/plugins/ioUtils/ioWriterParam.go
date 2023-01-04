package ioUtils

import (
	"errors"
	"fmt"
	"strings"
)

type _OUTPUT_TYPE int

const (
	_OUTPUT_FILE _OUTPUT_TYPE = 1
)

type ioWriterParam struct {
	OutputType _OUTPUT_TYPE
	FileOutput fileOutputParam
}

type fileOutputParam struct {
	OutFolder string // Folder to store output
}

func (ot *_OUTPUT_TYPE) UnmarshalJSON(b []byte) error {
	str := strings.Trim(string(b), `"`)

	switch {
	case str == "_OUTPUT_FILE":
		*ot = _OUTPUT_FILE
	default:
		return errors.New(fmt.Sprintf("Unknown option %s", str))
	}

	return nil
}
