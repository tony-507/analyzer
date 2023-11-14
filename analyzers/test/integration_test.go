package integration

// Integration test spec

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/tony-507/analyzers/src/common"
	"github.com/tony-507/analyzers/src/tttKernel"
	"github.com/tony-507/analyzers/test/schema"
	"github.com/tony-507/analyzers/test/validator"
)

func setupLogging(appDir string) {
	common.SetLoggingProperty("level", "trace")
	common.SetLoggingProperty("prefix", "[%l]")
	common.SetLoggingProperty("logDir", appDir+"logs")
}

func getOutputDir() string {
	_, filename, _, _ := runtime.Caller(0)
	return filepath.Dir(filename)
}

// TODO How to validate?
func TestListApp(t *testing.T) {
	tttKernel.ListApp("./resources/apps/")
}

func TestStartApp(t *testing.T) {
	specs := []string{
		"resources/testCases/ASCENT_tsa.json",
		"resources/testCases/ASCENT_editCap.json",
		"resources/testCases/AdSmart.json",
		"resources/testCases/AdSmart_pcap_tsa.json",
		"resources/testCases/EBPTimeCode_tsMon.json",
	}

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

		inFile := "resources/assets/" + tc.Source

		for _, app := range tc.App {
			outFolder := getOutputDir() + "/output/" + caseName + "/" + app + "/"
			noOutput := false
			testName := caseName + "_" + app
			t.Run(testName, func(t *testing.T) {
				setupLogging(outFolder)

				fmt.Println("Test:", testName)
				var args []string

				switch app {
				case "tsa":
					args = []string{
						"-f", inFile,
						"-o", outFolder,
					}
				case "editCap":
					args = []string{
						"-f", inFile,
						"-o", outFolder,
						"--maxInCnt", "50",
					}
				case "tsMon":
					args = []string{
						"-f1", inFile,
					}
					noOutput = true
				default:
					panic(fmt.Sprintf("Unknown app: %s", app))
				}

				// Run app
				fmt.Println("Running app")
				tttKernel.StartApp("./resources/apps/", app, args)

				if noOutput {
					return
				}

				// Move output files to outFolder
				entries, readErr := os.ReadDir("output")
				if readErr != nil {
					panic(readErr)
				}
				for _, file := range entries {
					if !file.IsDir() {
						os.Rename("output/"+file.Name(), outFolder+file.Name())
					}
				}

				// Perform validations
				fmt.Println("Performing validations")
				err := validator.PerformValidation(app, outFolder, tc.Expected)
				if err != nil {
					panic(err)
				}
			})
		}
	}
}
