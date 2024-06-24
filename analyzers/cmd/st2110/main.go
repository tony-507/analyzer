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

	flag.StringVar(&addresses, "addr", "", "Comma-separated list of URIs to monitor")
	flag.StringVar(&outDir, "o", "./output", "Output directory")
	flag.StringVar(&skipCnt, "skipCnt", "0", "Skip count")
	flag.StringVar(&maxInCnt, "maxInCnt", "0", "Max input packet count")

	flag.Parse()

	if addresses == "" {
		flag.Usage()
		return
	}

	builders := make([]tttKernel.OverallParams, 0)

	for idx, addr := range strings.Split(addresses, ",") {
		readerBuilder, err := controller.ReaderBuilder(&addr, idx)
		if err != nil {
			panic(err)
		}

		readerBuilder.SetProperty("Protocols", controller.NewProperty("TS"))
		readerBuilder.SetProperty("SkipCnt", controller.NewProperty(skipCnt))
		readerBuilder.SetProperty("MaxInCnt", controller.NewProperty(maxInCnt))

		bbProcBuilder := controller.NewPluginBuilder()
		bbProcBuilder.SetName(fmt.Sprintf("BasebandProcessor_%d", idx))
		bbProcBuilder.SetProperty("Mode", controller.NewProperty("_DEMUX_FULL"))

		controller.LinkPlugins([]*controller.PluginBuilder{
			&readerBuilder,
			&bbProcBuilder,
		})

		builders = append(builders, readerBuilder.Build())
		builders = append(builders, bbProcBuilder.Build())
	}

	controller.Start(
		&builders,
		&tttKernel.Resource{
			OutDir: outDir,
		},
	)
}
