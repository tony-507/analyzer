package integration

// Integration test spec

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"testing"
	"path/filepath"

	"github.com/tony-507/analyzers/src/tttKernel"
)

func TestIntegration(t *testing.T) {
	specs := []string{
		"resources/testCases/tsa_ASCENT.json",
	}

	// Debugging variables
	var curTest string
	var errMsg  string

	for _, spec := range specs {
		curTest = filepath.Base(spec)

		var tc testCase
		jsonString, err := ioutil.ReadFile(spec)
		if err != nil {
			errMsg = err.Error()
			break
		}

		err = json.Unmarshal([]byte(jsonString), &tc)
		if err != nil {
			errMsg = err.Error()
			break
		}

		for _, app := range tc.App {
			var args []string

			switch app {
			case "tsa":
				args = []string{
					"-f", "resources/assets/" + tc.Source,
					"-o", "output/" + filepath.Dir(tc.Source),
				}
			default:
				errMsg = fmt.Sprintf("Unknown app: %s", app)
				break
			}

			tttKernel.StartApp("./resources/apps/", app, args)
		}
	}

	if errMsg != "" {
		panic(fmt.Sprintf("[%s] Test fails due to %s", curTest, errMsg))
	}
}