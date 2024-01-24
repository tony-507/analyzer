package data

import (
	"errors"

	"github.com/tony-507/analyzers/src/plugins/common"
	"github.com/tony-507/analyzers/src/plugins/dataHandler/utils"
	"github.com/tony-507/analyzers/src/tttKernel"
)

type scte35Handler struct{}

func (h *scte35Handler) Feed(unit tttKernel.CmUnit, newData *utils.ParsedData) error {
	mediaUnit, ok := unit.(*common.MediaUnit)
	if !ok {
		return errors.New("Not a media unit")
	}

	data := newData.GetData()

	data.Type = utils.SCTE_35
	data.Scte35 = &utils.Scte35Struct{
		SpliceTime: int(mediaUnit.Data.GetField("spliceTime")),
	}
	return nil
}

func Scte35Handler(pid int) utils.DataHandler {
	return &scte35Handler{}
}
