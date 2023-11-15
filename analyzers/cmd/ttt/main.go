package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/tony-507/analyzers/src/common/logging"
	"github.com/tony-507/analyzers/src/controller"
)

func setupLogging(appDir string) {
	logging.SetLoggingProperty("level", "trace")
	logging.SetLoggingProperty("prefix", "[%l]")
	logging.SetLoggingProperty("logDir", appDir+"/ttt"+"_"+time.Now().Format("2006_01_02_15_04_05"))
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
	case "version":
		fmt.Println(controller.Version())
	case "ls":
		controller.ListApp(resourceDir)
	case "app":
		params := make([]string, 0)
		if len(os.Args) > 3 {
			params = os.Args[3:]
		}
		controller.StartApp(resourceDir, os.Args[2], params)
	default:
		showHelp()
	}
}
