package tttKernel

import (
	"strings"

	"github.com/tony-507/analyzers/src/common"
)

type IPlugin interface {
	SetCallback(common.RequestHandler)
	SetParameter(string)
	SetResource(*common.ResourceLoader)
	StartSequence()
	DeliverUnit(common.CmUnit, string)
	DeliverStatus(common.CmUnit)
	FetchUnit() common.CmUnit
	EndSequence()
	PrintInfo(*strings.Builder)
	Name() string
}
