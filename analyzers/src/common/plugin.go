package common

import "strings"

type IPlugin interface {
	SetCallback(RequestHandler)
	SetParameter(string)
	SetResource(*ResourceLoader)
	StartSequence()
	DeliverUnit(CmUnit, string)
	DeliverStatus(CmUnit)
	FetchUnit() CmUnit
	EndSequence()
	PrintInfo(*strings.Builder)
	Name() string
}
