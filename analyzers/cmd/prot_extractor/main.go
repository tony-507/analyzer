package main

import (
	"flag"

	"github.com/tony-507/analyzers/src/controller"
	"github.com/tony-507/analyzers/src/tttKernel"
)

func main() {
	var uri string
	var output string
	var protocols string

	flag.StringVar(&uri, "uri", "", "URI for extraction")
	flag.StringVar(&output, "output", "./output", "Output directory")
	flag.StringVar(&protocols, "protocols", "RTP", "Protocols to extract")

	flag.Parse()

	if uri == "" {
		flag.Usage()
		return
	}

	builders := make([]tttKernel.OverallParams, 0)

	readerBuilder := controller.NewPluginBuilder()
	readerBuilder.SetName("InputReader_1")
	readerBuilder.SetProperty("Uri", controller.NewProperty(uri))
	readerBuilder.SetProperty("Protocols", controller.NewProperty(protocols))
	readerBuilder.SetProperty("dumpRawInput", controller.NewProperty("true"))

	builders = append(builders, readerBuilder.Build())

	controller.LinkPlugins([]*controller.PluginBuilder{
		&readerBuilder,
	})

	controller.Start(
		&builders,
		&tttKernel.Resource{
			OutDir: output,
		},
	)
}