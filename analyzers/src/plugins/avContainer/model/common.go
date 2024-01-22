package model

import (
	"errors"
	"fmt"

	"github.com/tony-507/analyzers/src/tttKernel"
)

type DataStruct interface {
	Append(buf []byte)
	GetField(string) (int, error)
	GetName() string
	GetHeader() tttKernel.CmBuf
	GetPayload() []byte
	Ready() bool
	Serialize() []byte
	Process() error
}

type PsiManager interface {
	AddStream(version int, progNum int, streamPid int, streamType int)
	AddProgram(int, int, int)
	GetPATVersion() int
	GetPmtVersion(int) int
	GetPmtPidByProgNum(int) int
	PsiUpdateFinished(int, int, []byte)
	SpliceEventReceived(dpiPid int, spliceCmdTypeStr string, splicePTS []int, pktCnt int)
}

type pesHandle interface {
	PesPacketReady(buf tttKernel.CmBuf, pid int)
}

func resolveHeaderField(d DataStruct, str string) (int, error) {
	fieldStr, ok := d.GetHeader().GetField(str)
	if !ok {
		return 0, errors.New(fmt.Sprintf("%s does not exist in %s", str, d.GetName()))
	}
	rv, isInt := fieldStr.(int)
	if !isInt {
		return 0, errors.New(fmt.Sprintf("%s is not an integer in %s", str, d.GetName()))
	}
	return rv, nil
}
