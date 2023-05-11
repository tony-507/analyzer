package ioUtils

import (
	"errors"
	"fmt"
	"strings"
)

type _OUTPUT_TYPE int

type ioWriterParam struct {
	OutputType _OUTPUT_TYPE
}

func (ot *_OUTPUT_TYPE) UnmarshalJSON(b []byte) error {
	str := strings.Trim(string(b), `"`)

	return errors.New(fmt.Sprintf("Unknown option %s", str))
}
