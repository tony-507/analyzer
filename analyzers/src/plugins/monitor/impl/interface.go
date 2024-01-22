package impl

import (
	"errors"
	"fmt"
	"strings"

	"github.com/tony-507/analyzers/src/tttKernel"
)

type MonitorImpl interface {
	Feed(tttKernel.CmUnit, string)
	GetFields() []string
	GetDisplayData() []string
	HasInputId(string) bool
}

type redundancyTimeRef int

const (
	Pts redundancyTimeRef = 0
	Vitc                   = 1
)

func (st *redundancyTimeRef) UnmarshalJSON(b []byte) error {
	str := strings.Trim(string(b), `"`)
	switch str {
	case "pts":
		*st = Pts
	case "vitc":
		*st = Vitc
	default:
		return errors.New(fmt.Sprintf("Unknown redundancy time reference %s", str))
	}
	return nil
}

type RedundancyParam struct {
	TimeRef redundancyTimeRef
}