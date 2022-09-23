package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/tony-507/analyzers/src/controller"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Wrong number of arguments")
		fmt.Println("Usage: tsa <file>")
		os.Exit(1)
	}

	ex, _ := os.Executable()
	curDir := filepath.Dir(ex)
	fname := os.Args[1]

	sourceSetting := controller.SourceSetting{FileInput: controller.FileInputSetting{Fname: fname}}
	outputSetting := controller.OutputSetting{DataOutput: controller.DataOutputSetting{OutDir: curDir + "/output/" + strings.TrimSuffix(filepath.Base(fname), filepath.Ext(fname)) + "/"}}
	ctrlInterface := controller.CtrlInterface{SourceSetting: sourceSetting, OutputSetting: outputSetting}

	ctrl := controller.GetController(ctrlInterface)
	ctrl.StartApp()
}
