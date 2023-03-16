package common

type IPlugin interface {
	SetCallback(RequestHandler)
	SetParameter(string)
	SetResource(*ResourceLoader)
	StartSequence()
	DeliverUnit(CmUnit)
	DeliverStatus(CmUnit)
	FetchUnit() CmUnit
	EndSequence()
	IsRoot() bool
	Name() string
}
