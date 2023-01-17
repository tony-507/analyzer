package integration

// Integration test spec

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tony-507/analyzers/src/tttKernel"
	"github.com/tony-507/analyzers/test/schema"
	"github.com/tony-507/analyzers/test/validator"
)

func TestIntegration(t *testing.T) {
	specs := []string{
		"resources/testCases/ASCENT.json",
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
		outFolder := "output/" + caseName + "/"

		for _, app := range tc.App {
			testName := caseName + "_" + app
			t.Run(testName, func(t *testing.T) {
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
				default:
					panic(fmt.Sprintf("Unknown app: %s", app))
				}

				// Run app
				fmt.Println("Running app")
				tttKernel.StartApp("./resources/apps/", app, args)

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
