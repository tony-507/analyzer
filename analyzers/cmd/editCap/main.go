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

	flag.StringVar(&addr, "addr", "", "Pcap file to edit")
	flag.StringVar(&outDir, "o", "./output", "Output directory")
	flag.StringVar(&skipCnt, "skipCnt", "0", "Skip count")
	flag.StringVar(&maxInCnt, "maxInCnt", "0", "Max input packet count")

	flag.Parse()

	if addr == "" {
		flag.Usage()
		return
	}

	readerBuilder:= controller.NewPluginBuilder()
	readerBuilder.SetName("InputReader_0")
	readerBuilder.SetProperty("Uri", controller.NewProperty(addr))
	readerBuilder.SetProperty("Protocols", controller.NewProperty("TS"))
	readerBuilder.SetProperty("SkipCnt", controller.NewProperty(skipCnt))
	readerBuilder.SetProperty("MaxInCnt", controller.NewProperty(maxInCnt))
	readerBuilder.SetProperty("DumpRawInput", controller.NewProperty("true"))

	controller.Start(
		&[]tttKernel.OverallParams{readerBuilder.Build()},
		&tttKernel.Resource{
			OutDir: outDir,
		},
	)
}
