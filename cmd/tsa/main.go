package main

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/tony-507/analyzers/src/controller"
	"github.com/tony-507/analyzers/src/logs"
)

func setupLogging() {
	logs.SetProperty("level", "info")
	logs.SetProperty("format", "%t [%n] [%l] %s")
}

func main() {
	setupLogging()
	logger := logs.CreateLogger("main")
	if len(os.Args) != 2 {
		logger.Log(logs.ERROR, "Wrong number of arguments")
		logger.Log(logs.INFO, "Usage: tsa <file>")
		os.Exit(1)
	}

	ex, _ := os.Executable()
	curDir := filepath.Dir(ex)
	fname := os.Args[1]

	sourceSetting := controller.SourceSetting{FileInput: controller.FileInputSetting{Fname: fname}}
	demuxSetting := controller.DemuxSetting{Mode: 2}
	outputSetting := controller.OutputSetting{DataOutput: controller.DataOutputSetting{OutDir: curDir + "/output/" + strings.TrimSuffix(filepath.Base(fname), filepath.Ext(fname)) + "/"}}
	ctrlInterface := controller.CtrlInterface{SourceSetting: sourceSetting, DemuxSetting: demuxSetting, OutputSetting: outputSetting}

	ctrl := controller.GetController(ctrlInterface)
	ctrl.StartApp()
}
