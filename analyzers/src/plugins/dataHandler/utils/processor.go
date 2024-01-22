package utils

import (
	"strings"

	"github.com/tony-507/analyzers/src/tttKernel"
)

type DataProcessor interface {
	PrintInfo(*strings.Builder)
	Start() error
	Stop() error
	Process(tttKernel.CmUnit, *ParsedData)
}