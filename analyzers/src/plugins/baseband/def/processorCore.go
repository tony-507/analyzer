package def

import (
	"strings"

	"github.com/tony-507/analyzers/src/common"
	"github.com/tony-507/analyzers/src/tttKernel"
)

type ProcessorCallback interface {
	OnDataReady(tttKernel.CmUnit)     // Ready to be fetched
}

type ProcessorCore interface {
	DeliverData(*common.MediaUnit)
	Feed(tttKernel.CmUnit, string)    // Feed a unit to core
	PrintInfo(sb *strings.Builder) // Periodic debug info
	SetCallback(ProcessorCallback) // Set plugin callback for OnDataReady
}
