package def

import (
	"strings"

	"github.com/tony-507/analyzers/src/common"
)

type ProcessorCallback interface {
	OnDataReady(common.CmUnit)     // Ready to be fetched
}

type ProcessorCore interface {
	DeliverData(*common.MediaUnit)
	Feed(common.CmUnit, string)    // Feed a unit to core
	PrintInfo(sb *strings.Builder) // Periodic debug info
	SetCallback(ProcessorCallback) // Set plugin callback for OnDataReady
}
