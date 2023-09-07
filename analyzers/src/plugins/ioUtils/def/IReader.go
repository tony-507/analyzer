package def

import "github.com/tony-507/analyzers/src/common"

type IReader interface {
	Setup(config IReaderConfig)        // Set up reader
	StartRecv() error                  // Start receiver
	StopRecv() error                   // Stop receiver
	DataAvailable(*common.IOUnit) bool // Get next unit of data
}

type IReaderConfig struct {
	Protocols []PROTOCOL
}