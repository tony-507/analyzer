package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/tony-507/analyzers/src/controller"
	"github.com/tony-507/analyzers/src/logs"
)

var resourceDir string

func setupLogging() {
	logs.SetProperty("level", "trace")
	logs.SetProperty("prefix", "[%l]")
}

func showHelp() {
	fmt.Println("Usage: ttt <appName> <parameters>...")
}

func getApps(appName string) string {
	fileInfo, err := ioutil.ReadDir(resourceDir)
	if err != nil {
		panic(err)
	}
	rv := ""

	for _, file := range fileInfo {
		app := strings.Split(file.Name(), ".")[0]
		if app == appName {
			buf, err := ioutil.ReadFile(resourceDir + file.Name())
			if err != nil {
				panic(err)
			}
			rv = string(buf)
		}
	}

	return rv
}

func main() {
	setupLogging()
	ex, _ := os.Executable()
	resourceDir = filepath.Dir(ex) + "/.resources/"

	if len(os.Args) < 2 {
		showHelp()
		return
	}

	switch os.Args[1] {
	case "help":
		showHelp()
	default:
		script := getApps(os.Args[1])
		params := make([]string, 0)
		if len(os.Args) > 2 {
			params = os.Args[2:]
		}
		controller.GetMigratedController(script, params)
	}
}
