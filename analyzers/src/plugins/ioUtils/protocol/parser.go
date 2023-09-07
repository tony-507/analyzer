package protocol

import (
	"fmt"

	"github.com/tony-507/analyzers/src/plugins/ioUtils/def"
)

func ParseWithParsers(protocols []def.PROTOCOL, rawBuf []byte) []def.ParseResult {
	if len(protocols) == 1 {
		return parseOnce(protocols[0], rawBuf)
	}

	prot := protocols[0]
	res := []def.ParseResult{}
	for _, item := range(parseOnce(prot, rawBuf)) {
		// TODO: Do not discard info from other protocols
		res = append(res, ParseWithParsers(protocols[1:], item.GetBuffer())...)
	}
	return res
}

func parseOnce(protocol def.PROTOCOL, rawBuf []byte) []def.ParseResult {
	var parser def.IParser
	switch protocol {
	case def.PROT_TS:
		parser = TsParser()
	default:
		panic(fmt.Sprintf("Parser %v not supported", protocol))
	}
	return parser.Parse(rawBuf)
}