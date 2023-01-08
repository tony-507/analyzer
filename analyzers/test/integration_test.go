package integration

// Integration test spec

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"
	"testing"
	"path/filepath"

	"github.com/tony-507/analyzers/src/tttKernel"
	"github.com/tony-507/analyzers/test/schema"
	"github.com/tony-507/analyzers/test/validator"
)

func TestIntegration(t *testing.T) {
	specs := []string{
		"resources/testCases/tsa_ASCENT.json",
	}

	for _, spec := range specs {
		testName := strings.Split(filepath.Base(spec), ".")[0]

		t.Run(testName, func(t *testing.T) {
			var tc schema.TestCase

			// Preparation
			fmt.Println("Initializing test")
			jsonString, err := ioutil.ReadFile(spec)
			if err != nil {
				panic(err)
			}

			err = json.Unmarshal([]byte(jsonString), &tc)
			if err != nil {
				panic(err)
			}

			inFile := "resources/assets/" + tc.Source
			outFolder := "output/" + testName + "/"

			for _, app := range tc.App {
				var args []string

				switch app {
				case "tsa":
					args = []string{
						"-f", inFile,
						"-o", outFolder,
					}
				default:
					panic(fmt.Sprintf("Unknown app: %s", app))
				}

				// Run app
				fmt.Println("Running app")
				tttKernel.StartApp("./resources/apps/", app, args)

				// Perform validations
				fmt.Println("Performing validations")
				err := validator.PerformValidation(outFolder, tc.Expected)
				if err != nil {
					panic(err)
				}
			}
		})
	}
}