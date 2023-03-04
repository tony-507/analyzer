package model

import (
	"fmt"
	"errors"

	"github.com/tony-507/analyzers/src/common"
)

type DataStruct interface {
	Append(buf []byte)
	GetField(string)   (int, error)
	GetName()          string
	getHeader()        common.CmBuf
	GetPayload()       []byte
	Ready()            bool
	Serialize()        []byte
}

func resolveHeaderField(d DataStruct, str string) (int, error) {
	fieldStr, ok := d.getHeader().GetField(str)
	if !ok {
		return 0, errors.New(fmt.Sprintf("%s does not exist in %s", str, d.GetName()))
	}
	rv, isInt := fieldStr.(int)
	if !isInt {
		return 0, errors.New(fmt.Sprintf("%s is not an integer in %s", str, d.GetName()))
	}
	return rv, nil
}