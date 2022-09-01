package main

import (
	"avContainer"
	"ioUtils"
)

// Maybe use a wrapper to make them all implement same interface?

func main() {
	inputReader := ioUtils.GetReader("C:\\Users\\tchan\\Desktop\\proj\\Jira\\NG-67963\\ASI_out.ts")
	demuxer := avContainer.GetTsDemuxer()
	writer := ioUtils.GetFileWriter("D:\\workspace\\analyzers\\output\\NG-67963_ASI\\")

	for {
		if inputReader.DataAvailable() {
			unit := inputReader.Feed()
			demuxer.DeliverUnit(unit)
		} else {
			break
		}
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
