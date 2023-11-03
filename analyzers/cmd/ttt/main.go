package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/tony-507/analyzers/src/common"
	"github.com/tony-507/analyzers/src/tttKernel"
)

func setupLogging(appDir string) {
	common.SetLoggingProperty("level", "trace")
	common.SetLoggingProperty("prefix", "[%l]")
	common.SetLoggingProperty("logDir", appDir+"/ttt"+"_"+time.Now().Format("2006_01_02_15_04_05"))
}

func showHelp() {
	fmt.Println("Usage: ttt <appName> <parameters>...")
}

func main() {
	ex, _ := os.Executable()
	appDir := filepath.Dir(ex)
	resourceDir := appDir + "/.resources/"
	setupLogging(appDir)

	if len(os.Args) < 2 {
		showHelp()
		return
	}

	switch os.Args[1] {
	case "help":
		showHelp()
	case "ls":
		tttKernel.ListApp(resourceDir)
	case "app":
		params := make([]string, 0)
		if len(os.Args) > 3 {
			params = os.Args[3:]
		}
		tttKernel.StartApp(resourceDir, os.Args[2], params)
	default:
		showHelp()
	}
}
