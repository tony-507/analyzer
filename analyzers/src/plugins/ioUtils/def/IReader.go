package def

type IReader interface {
	Setup(config IReaderConfig)         // Set up reader
	StartRecv() error                   // Start receiver
	StopRecv() error                    // Stop receiver
	DataAvailable() (ParseResult, bool) // Get next unit of data
}

type IReaderConfig struct {
	Protocols []PROTOCOL
}