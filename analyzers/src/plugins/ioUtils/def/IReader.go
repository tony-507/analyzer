package def

import "github.com/tony-507/analyzers/src/plugins/common/protocol"

type IReader interface {
	Setup(config IReaderConfig)              // Set up reader
	StartRecv() error                        // Start receiver
	StopRecv() error                         // Stop receiver
	DataAvailable() (protocol.ParseResult, bool) // Get next unit of data
}

type IReaderConfig struct {
	Parsers []protocol.IParser
}