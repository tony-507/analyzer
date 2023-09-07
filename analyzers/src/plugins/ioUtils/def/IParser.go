package def

type IParser interface {
	Parse([]byte) []ParseResult // Parse given data
}

type ParseResult interface {
	// Get data
	GetBuffer() []byte
	// Get a parsed field as numerical value. Return false in second field if not exist
	GetField(string) (int, bool)
}