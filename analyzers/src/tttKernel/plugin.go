package tttKernel

import (
	"strings"

	"github.com/tony-507/analyzers/src/common"
)

type IPlugin interface {
	SetCallback(RequestHandler)
	SetParameter(string)
	SetResource(*ResourceLoader)
	StartSequence()
	DeliverUnit(common.CmUnit, string)
	DeliverStatus(common.CmUnit)
	FetchUnit() common.CmUnit
	EndSequence()
	PrintInfo(*strings.Builder)
	Name() string
}
