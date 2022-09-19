package datahandler

import (
	"fmt"

	"github.com/tony-507/analyzers/src/common"
)

type DataHandler interface {
	Feed(buf []byte) // Accept input buffer and begin parsing
}

type DataHandlerFactory struct {
	handlers   map[int]int
	outputUnit []common.CmUnit
	isRunning  bool
	name       string
}

func (df *DataHandlerFactory) SetParameter(m_parameter interface{}) {
	df._setup()
}

func (df *DataHandlerFactory) _setup() {
	df.handlers = make(map[int]int, 0)
	df.outputUnit = make([]common.CmUnit, 0)
	df.isRunning = true
}

func (df *DataHandlerFactory) StopPlugin() {
	fmt.Println("Data handler factory is stopped")
}

// Handle only PesBuf
func (df *DataHandlerFactory) DeliverUnit(unit common.CmUnit) common.CmUnit {
	if unit == nil {
		return nil
	}

	// Extract buffer from input unit
	_, isPes := unit.GetBuf().(common.PesBuf)
	if isPes {
		pid, isPidInt := unit.GetField("id").(int)
		if !isPidInt {
			panic("Something wrong with the data")
		}

		_, hasPid := df.handlers[pid]
		if !hasPid {
			fmt.Println("Receive pid", pid, "at dataHandlerFactory")
			df.handlers[pid] = 1
		}
	}

	df.outputUnit = append(df.outputUnit, unit)

	// Directly output the unit
	reqUnit := common.MakeReqUnit(nil, common.FETCH_REQUEST)
	return reqUnit
}

func (df *DataHandlerFactory) FetchUnit() common.CmUnit {
	if len(df.outputUnit) == 0 {
		return nil
	} else if len(df.outputUnit) == 1 {
		rv := df.outputUnit[0]
		df.outputUnit = make([]common.CmUnit, 0)
		return rv
	} else {
		rv := df.outputUnit[0]
		df.outputUnit = df.outputUnit[1:]
		return rv
	}
}

func GetDataHandlerFactory(name string) DataHandlerFactory {
	rv := DataHandlerFactory{name: name}
	return rv
}
