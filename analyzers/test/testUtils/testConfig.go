package testUtils

import (
	"os"
	"path/filepath"
)

func GetOutputDir() string {
	ex, _ := os.Executable()
	return filepath.Dir(ex)
}
