package def

import "fmt"

type IParser interface {
	Parse([]byte) []ParseResult // Parse given data
}

type ParseResult struct {
	Buffer []byte
	Fields map[string]int64
}

func (res *ParseResult) GetBuffer() []byte {
	return res.Buffer
}

func (res *ParseResult) GetField(name string) (int64, bool) {
	val, ok := res.Fields[name]
	return val, ok
}

func AssertIntEqual(name string, expected int, actual int) {
	if expected != actual {
		panic(fmt.Sprintf("Invalid value at %s. Expecting %d but got %d", name, expected, actual))
	}
}