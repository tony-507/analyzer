package utils

import "github.com/tony-507/analyzers/src/common"

type DataHandler interface {
	Feed(unit common.CmUnit) // Accept input buffer and begin parsing
}
