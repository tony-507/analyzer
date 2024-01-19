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

func validateFileProp(outFolder string, expected *[]schema.FileProp) error {
	fileInfo, err := ioutil.ReadDir(outFolder)
	for _, prop := range *expected {
		size := -1
		if prop.Size != 0 {
			size = prop.Size
		}

		if err = hasNonEmptyFileInList(prop.Fname, size, fileInfo); err != nil {
			return err
		}

		if prop.Md5sum != "" {
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
