package validator

import (
	"fmt"
	"io/fs"
	"io/ioutil"
	"errors"

	"github.com/tony-507/analyzers/test/schema"
)

func PerformValidation(outFolder string, expectedProp *schema.ExpectedProp) error {
	if expectedProp != nil {
		if expectedProp.Tsa != nil {
			err := validateTsa(outFolder, expectedProp.Tsa)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func validateTsa(outFolder string, expectedTsaProp *schema.TsaExpectedProp) error {
	fileInfo, err := ioutil.ReadDir(outFolder)
	if err != nil {
		panic(err)
	}

	for _, fname := range expectedTsaProp.CsvList {
		if !hasNonEmptyFileInList(fname + ".csv", fileInfo) {
			return errors.New(fmt.Sprintf("%s.csv not found in output folder", fname))
		}
	}

	for _, fname := range expectedTsaProp.JsonList {
		if !hasNonEmptyFileInList(fname + ".json", fileInfo) {
			return errors.New(fmt.Sprintf("%s.json not found in output folder", fname))
		}
	}

	return nil
}

func hasNonEmptyFileInList(fname string, fileInfo []fs.FileInfo) bool {
	for _, file := range fileInfo {
		if file.Name() == fname && file.Size() != 0 {
			return true
		}
	}
	return false
}