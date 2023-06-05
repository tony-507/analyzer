package validator

import (
	"errors"
	"fmt"
	"io/fs"
	"io/ioutil"
	"os/exec"
	"strings"

	"github.com/tony-507/analyzers/test/schema"
)

func PerformValidation(app string, outFolder string, expectedProp *schema.ExpectedProp) error {
	if expectedProp != nil {
		switch app {
		case "tsa":
			if expectedProp.Tsa != nil {
				err := validateTsa(outFolder, expectedProp.Tsa)
				if err != nil {
					return err
				}
			}
		case "editCap":
			if expectedProp.EditCap != nil {
				err := validateEditCap(outFolder, expectedProp.EditCap)
				if err != nil {
					return err
				}
			}
		default:
			return errors.New(fmt.Sprintf("Unknown app %s", app))
		}
	}
	return nil
}

func validateTsa(outFolder string, expectedTsaProp *schema.TsaExpectedProp) error {
	fileInfo, err := ioutil.ReadDir(outFolder)
	if err != nil {
		return err
	}

	for _, fname := range expectedTsaProp.CsvList {
		if err = hasNonEmptyFileInList(fname+".csv", -1, fileInfo); err != nil {
			return err
		}
	}

	for _, fname := range expectedTsaProp.JsonList {
		if err = hasNonEmptyFileInList(fname+".json", -1, fileInfo); err != nil {
			return err
		}
	}

	if expectedTsaProp.Md5sum != nil {
		for _, prop := range *expectedTsaProp.Md5sum {
			out, err := exec.Command("md5sum", fmt.Sprintf("%s%s", outFolder, prop.Fname)).Output()
			if err != nil {
				return err
			}
			actualMd5sum := strings.Split(string(out), " ")[0]
			if actualMd5sum != prop.Md5sum {
				return errors.New(fmt.Sprintf("Md5sum not match. Expecting %s, but got %s", prop.Md5sum, actualMd5sum))
			}
		}
	}

	return nil
}

func validateEditCap(outFolder string, expectedEditCapProp *schema.EditCapExpectedProp) error {
	fileInfo, err := ioutil.ReadDir(outFolder)
	if err != nil {
		return err
	}

	if err = hasNonEmptyFileInList(expectedEditCapProp.Fname, expectedEditCapProp.Size, fileInfo); err != nil {
		return err
	}
	return nil
}

func hasNonEmptyFileInList(fname string, size int, fileInfo []fs.FileInfo) error {
	errMsg := fmt.Sprintf("%s not found in output folder", fname)
	for _, file := range fileInfo {
		if file.Name() == fname {
			if size != -1 {
				if file.Size() == int64(size) {
					return nil
				} else {
					errMsg = fmt.Sprintf("File size of %s is not %d but %d", fname, size, file.Size())
				}
			} else {
				if file.Size() > 20 {
					return nil
				} else {
					errMsg = fmt.Sprintf("%s is empty", fname)
				}
			}
		}
	}
	return errors.New(errMsg)
}
