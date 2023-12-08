package protocol

import (
	"fmt"
)

func ParseWithParsers(parsers []IParser, lastRes *ParseResult) []ParseResult {
	if len(parsers) == 0 {
		return []ParseResult{*lastRes}
	}

	if len(parsers) == 1 {
		return parsers[0].Parse(lastRes)
	}

	parser := parsers[0]
	res := []ParseResult{}
	for _, item := range(parser.Parse(lastRes)) {
		// TODO: Do not discard info from other protocols
		res = append(res, ParseWithParsers(parsers[1:], &item)...)
	}
	return res
}

func GetParser(protocol PROTOCOL) IParser {
	switch protocol {
	case PROT_TS:
		return TsParser()
	case PROT_RTP:
		return RtpParser()
	default:
		panic(fmt.Sprintf("Parser %v not supported", protocol))
	}
}
