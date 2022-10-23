package main

import (
	"os"
	"path/filepath"
	"strings"
	"strconv"

	"github.com/tony-507/analyzers/src/controller"
	"github.com/tony-507/analyzers/src/logs"
)

func main() {
	logger := logs.CreateLogger("main")
	if len(os.Args) != 4 {
		logger.Log(logs.ERROR, "Wrong number of arguments")
		logger.Log(logs.INFO, "Usage: editCap <start> <end>")
		os.Exit(1)
	}

	ex, _ := os.Executable()
	curDir := filepath.Dir(ex)

	fname := os.Args[1]
	skipCnt, _ := strconv.Atoi(os.Args[2])
	maxInCnt, _ := strconv.Atoi(os.Args[3])
	sourceSetting := controller.SourceSetting{FileInput: controller.FileInputSetting{Fname: fname},
		SkipCnt: skipCnt, MaxInCnt: maxInCnt}
	outputSetting := controller.OutputSetting{DataOutput: controller.DataOutputSetting{OutDir: curDir + "/output/" + strings.TrimSuffix(filepath.Base(fname), filepath.Ext(fname)) + "/"}}
	ctrlInterface := controller.CtrlInterface{SourceSetting: sourceSetting, OutputSetting: outputSetting}

	ctrl := controller.GetController(ctrlInterface)
	ctrl.StartApp()
}
