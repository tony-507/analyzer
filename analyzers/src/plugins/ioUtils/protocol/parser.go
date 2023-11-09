package protocol

import (
	"fmt"

	"github.com/tony-507/analyzers/src/plugins/ioUtils/def"
)

func ParseWithParsers(parsers []def.IParser, lastRes *def.ParseResult) []def.ParseResult {
	if len(parsers) == 1 {
		return parsers[0].Parse(lastRes)
	}

	parser := parsers[0]
	res := []def.ParseResult{}
	for _, item := range(parser.Parse(lastRes)) {
		// TODO: Do not discard info from other protocols
		res = append(res, ParseWithParsers(parsers[1:], &item)...)
	}
	return res
}

func GetParser(protocol def.PROTOCOL) def.IParser {
	switch protocol {
	case def.PROT_TS:
		return TsParser()
	case def.PROT_RTP:
		return RtpParser()
	default:
		panic(fmt.Sprintf("Parser %v not supported", protocol))
	}
}
