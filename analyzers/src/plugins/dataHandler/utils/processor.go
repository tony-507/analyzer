package utils

import (
	"strings"

	"github.com/tony-507/analyzers/src/common"
)

type DataProcessor interface {
	PrintInfo(*strings.Builder)
	Start() error
	Stop() error
	Process(common.CmUnit, *ParsedData)
}