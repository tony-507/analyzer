package impl

import "github.com/tony-507/analyzers/src/common"

type MonitorImpl interface {
	Feed(common.CmUnit, string)
	GetFields() []string
	GetDisplayData() []string
	HasInputId(string) bool
}
