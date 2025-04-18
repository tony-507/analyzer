package main

import (
	"flag"
	"fmt"
	"strings"

	"github.com/tony-507/analyzers/src/controller"
	"github.com/tony-507/analyzers/src/tttKernel"
)

func main() {
	var addresses string
	var outDir string
	var skipCnt string
	var maxInCnt string
	var redundancy string

	flag.StringVar(&addresses, "addr", "", "Comma-separated list of URIs to monitor")
	flag.StringVar(&outDir, "o", "./output", "Output directory")
	flag.StringVar(&skipCnt, "skipCnt", "0", "Skip count")
	flag.StringVar(&maxInCnt, "maxInCnt", "0", "Max input packet count")
	flag.StringVar(&redundancy, "redundancy", "None", "Redundancy time reference")

	flag.Parse()

	if addresses == "" {
		flag.Usage()
		return
	}

	builders := make([]tttKernel.OverallParams, 0)

	monitorBuilder := controller.NewPluginBuilder()
	monitorBuilder.SetName("OutputMonitor_0")
	if redundancy != "None" {
		monitorBuilder.SetProperty("Redundancy.TimeRef", controller.NewProperty(redundancy))
	}

	builders = append(builders, monitorBuilder.Build())

	for idx, addr := range strings.Split(addresses, ",") {
		readerBuilder := controller.NewPluginBuilder()
		readerBuilder.SetName(fmt.Sprintf("InputReader_%d", idx))
		readerBuilder.SetProperty("Uri", controller.NewProperty(addr))
		readerBuilder.SetProperty("Protocols", controller.NewProperty("TS"))
		readerBuilder.SetProperty("SkipCnt", controller.NewProperty(skipCnt))
		readerBuilder.SetProperty("MaxInCnt", controller.NewProperty(maxInCnt))

		demuxBuilder := controller.NewPluginBuilder()
		demuxBuilder.SetName(fmt.Sprintf("TsDemuxer_%d", idx))
		demuxBuilder.SetProperty("Mode", controller.NewProperty("_DEMUX_FULL"))

		dataHdlrBuilder := controller.NewPluginBuilder()
		dataHdlrBuilder.SetName(fmt.Sprintf("DataHandler_%d", idx))

		controller.LinkPlugins([]*controller.PluginBuilder{
			&readerBuilder,
			&demuxBuilder,
			&dataHdlrBuilder,
			&monitorBuilder,
		})

		builders = append(builders, readerBuilder.Build())
		builders = append(builders, demuxBuilder.Build())
		builders = append(builders, dataHdlrBuilder.Build())
	}

	controller.Start(
		&builders,
		&tttKernel.Resource{
			OutDir: outDir,
		},
	)
}
