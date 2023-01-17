package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/tony-507/analyzers/src/logs"
	"github.com/tony-507/analyzers/src/tttKernel"
)

func setupLogging(appDir string) {
	logs.SetProperty("level", "trace")
	logs.SetProperty("prefix", "[%l]")
	logs.SetProperty("logDir", appDir+"/ttt"+"_"+strconv.Itoa(int(time.Now().Unix())))
}

func showHelp() {
	fmt.Println("Usage: ttt <appName> <parameters>...")
}

func main() {
	ex, _ := os.Executable()
	appDir := filepath.Dir(ex)
	setupLogging(appDir)

	if len(os.Args) < 2 {
		showHelp()
		return
	}

	switch os.Args[1] {
	case "help":
		showHelp()
	default:
		params := make([]string, 0)
		if len(os.Args) > 2 {
			params = os.Args[2:]
		}
		tttKernel.StartApp(appDir+"/.resources/", os.Args[1], params)
	}
}
