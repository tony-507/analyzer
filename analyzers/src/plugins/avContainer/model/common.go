package model

import (
	"fmt"
	"errors"

	"github.com/tony-507/analyzers/src/common"
)

type baseInterface interface {
	Append(buf []byte)
	GetField(string)   (int, error)
	GetName()          string
	getHeader()        common.CmBuf
	GetPayload()       []byte
	Ready()            bool
	Serialize()        []byte
}

type psiInterface interface {
	ParsePayload() error
}

type PsiManager interface {
	AddStream(version int, progNum int, streamPid int, streamType int)
	AddProgram(int, int, int)
	GetPATVersion()         int
	GetPmtVersion(int)      int
	GetPmtPidByProgNum(int) int
	PsiUpdateFinished(int, []byte)
	SpliceEventReceived(dpiPid int, spliceCmdTypeStr string, splicePTS []int)
}

type DataStruct interface {
	baseInterface
}

type TableStruct interface {
	baseInterface
	psiInterface
}

func resolveHeaderField(d baseInterface, str string) (int, error) {
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