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
)

func TestIntegration(t *testing.T) {
	specs := []string{
		"resources/testCases/tsa_ASCENT.json",
	}

	for _, spec := range specs {
		testName := strings.Split(filepath.Base(spec), ".")[0]

		t.Run(testName, func(t *testing.T) {
			var tc testCase
			jsonString, err := ioutil.ReadFile(spec)
			if err != nil {
				panic(err)
			}

			err = json.Unmarshal([]byte(jsonString), &tc)
			if err != nil {
				panic(err)
			}

			for _, app := range tc.App {
				var args []string

				switch app {
				case "tsa":
					args = []string{
						"-f", "resources/assets/" + tc.Source,
						"-o", "output/" + testName + "/",
					}
				default:
					panic(fmt.Sprintf("Unknown app: %s", app))
				}

				tttKernel.StartApp("./resources/apps/", app, args)
			}
		})
	}
}