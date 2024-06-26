package integration_test

// Integration test spec

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/tony-507/analyzers/src/logging"
	"github.com/tony-507/analyzers/test/schema"
	"github.com/tony-507/analyzers/test/validator"
)

func setupLogging(appDir string) {
	logging.SetLoggingProperty("level", "trace")
	logging.SetLoggingProperty("prefix", "[%l]")
	logging.SetLoggingProperty("logDir", appDir+"logs")
}

func getOutputDir() string {
	_, filename, _, _ := runtime.Caller(0)
	return filepath.Dir(filename)
}

func readTestCases(testCaseDir string) []string {
	specs := []string{}

	fileInfo, err := ioutil.ReadDir(testCaseDir)
	if err != nil {
		panic(fmt.Sprintf("Fail to read testcases: %s", err.Error()))
	}

	for _, file := range fileInfo {
		specs = append(specs, filepath.Join(testCaseDir, file.Name()))
	}

	return specs
}

func TestStartApp(t *testing.T) {
	specs := readTestCases("resources/testCases")

	for _, spec := range specs {
		caseName := strings.Split(filepath.Base(spec), ".")[0]
		var tc schema.TestCase

		// Preparation
		fmt.Println("Initializing test for", caseName)
		jsonString, err := ioutil.ReadFile(spec)
		if err != nil {
			panic(err)
		}

		err = json.Unmarshal([]byte(jsonString), &tc)
		if err != nil {
			panic(err)
		}

		inFile := filepath.Join(getOutputDir(), "resources/assets/", tc.Source)

		for _, app := range tc.App {
			outFolder := getOutputDir() + "/output/" + caseName + "/" + app + "/"
			testName := caseName + "_" + app
			t.Run(testName, func(t *testing.T) {
				setupLogging(outFolder)

				fmt.Println("Test:", testName)
				var args []string

				switch app {
				case "tsa":
					args = []string{
						"-addr", fmt.Sprintf("file://%s", inFile),
						"-o", outFolder,
					}
				case "editCap":
					args = []string{
						"-addr", fmt.Sprintf("file://%s", inFile),
						"-o", outFolder,
						"-maxInCnt", "50",
					}
				case "tsMonitor":
					args = []string{
						"-addr", fmt.Sprintf("file://%s", inFile),
						"-o", outFolder,
					}
				default:
					panic(fmt.Sprintf("Unknown app: %s", app))
				}

				// Run app
				fmt.Println("Running app")
				cmd := exec.Command(
					fmt.Sprintf("./../build/bin/%s", app),
					args...,
				)

				err := cmd.Run()
				if err != nil {
					panic(err)
				}

				// Perform validations
				fmt.Println("Performing validations")
				err = validator.PerformValidation(app, outFolder, tc.Expected)
				if err != nil {
					panic(err)
				}
			})
		}
	}
}
