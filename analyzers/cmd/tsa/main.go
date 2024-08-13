package main

import (
	"flag"

	"github.com/tony-507/analyzers/src/controller"
	"github.com/tony-507/analyzers/src/tttKernel"
)

func main() {
	var addr string
	var outDir string
	var skipCnt string
	var maxInCnt string

	flag.StringVar(&addr, "addr", "", "URI to analyze")
	flag.StringVar(&outDir, "o", "./output", "Output directory")
	flag.StringVar(&skipCnt, "skipCnt", "0", "Skip count")
	flag.StringVar(&maxInCnt, "maxInCnt", "0", "Max input packet count")

	flag.Parse()

	if addr == "" {
		flag.Usage()
		return
	}

	builders := make([]tttKernel.OverallParams, 0)

	readerBuilder := controller.NewPluginBuilder()
	readerBuilder.SetName("InputReader_0")
	readerBuilder.SetProperty("Uri", controller.NewProperty(addr))
	readerBuilder.SetProperty("Protocols", controller.NewProperty("TS"))
	readerBuilder.SetProperty("SkipCnt", controller.NewProperty(skipCnt))
	readerBuilder.SetProperty("MaxInCnt", controller.NewProperty(maxInCnt))

	demuxBuilder := controller.NewPluginBuilder()
	demuxBuilder.SetName("TsDemuxer_0")
	demuxBuilder.SetProperty("Mode", controller.NewProperty("_DEMUX_FULL"))

	dataHdlrBuilder := controller.NewPluginBuilder()
	dataHdlrBuilder.SetName("DataHandler_0")

	controller.LinkPlugins([]*controller.PluginBuilder{
		&readerBuilder,
		&demuxBuilder,
		&dataHdlrBuilder,
	})

	builders = append(builders, readerBuilder.Build())
	builders = append(builders, demuxBuilder.Build())
	builders = append(builders, dataHdlrBuilder.Build())

	controller.Start(
		&builders,
		&tttKernel.Resource{
			OutDir: outDir,
		},
	)
}
