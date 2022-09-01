package main

import (
	"github.com/tony-507/analyzers/src/avContainer"
	"github.com/tony-507/analyzers/src/common"
	"github.com/tony-507/analyzers/src/ioUtils"
)

// Maybe use a wrapper to make them all implement same interface?

func main() {
	inputReader := ioUtils.GetReader()
	inputReader.SetFileHandle("C:\\Users\\tchan\\Desktop\\proj\\Jira\\NG-67963\\ASI_out.ts")
	demuxer := avContainer.GetTsDemuxer()
	writer := ioUtils.GetFileWriter()
	writer.SetFolder("D:\\workspace\\analyzers\\output\\NG-67963_ASI\\")

	dummy := common.IOUnit{}

	for {
		if inputReader.DataAvailable() {
			inputReader.DeliverUnit(dummy)
		} else {
			break
		}
	}

	for {
		unit := inputReader.FetchUnit()
		if unit.GetField("type") == 0 {
			break
		}
		demuxer.DeliverUnit(unit)
	}

	for {
		unit := demuxer.FetchUnit()
		if unit.GetField("type") == 0 {
			break
		}
		writer.DeliverUnit(unit)
	}

	demuxer.StopPlugin()
	writer.StopPlugin()
}
