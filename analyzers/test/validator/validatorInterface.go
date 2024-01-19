package validator

import (
	"fmt"

	"github.com/tony-507/analyzers/test/schema"
)

func PerformValidation(app string, outFolder string, expectedProp *schema.ExpectedProp) error {
	if expectedProp == nil {
		return nil
	}

	var prop *schema.TestProp = nil
	switch app {
	case "tsa":
		prop = expectedProp.Tsa
	case "editCap":
		prop = expectedProp.EditCap
	default:
		fmt.Println(fmt.Sprintf("No validation done for app %s", app))
	}

	if prop == nil {
		return nil
	}

	if prop.File != nil {
		return validateFileProp(outFolder, prop.File)
	}

	return nil
}
