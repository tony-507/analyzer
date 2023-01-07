package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/tony-507/analyzers/src/logs"
	"github.com/tony-507/analyzers/src/tttKernel"
)

func setupLogging() {
	logs.SetProperty("level", "trace")
	logs.SetProperty("prefix", "[%l]")
}

func showHelp() {
	fmt.Println("Usage: ttt <appName> <parameters>...")
}

func main() {
	setupLogging()
	ex, _ := os.Executable()
	resourceDir := filepath.Dir(ex) + "/.resources/"

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
		tttKernel.StartApp(resourceDir, os.Args[1], params)
	}
}
