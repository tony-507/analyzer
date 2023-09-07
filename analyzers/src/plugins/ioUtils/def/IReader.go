package def

import "github.com/tony-507/analyzers/src/common"

type IReader interface {
	Setup()                            // Set up reader
	StartRecv() error                  // Start receiver
	StopRecv() error                   // Stop receiver
	DataAvailable(*common.IOUnit) bool // Get next unit of data
}
