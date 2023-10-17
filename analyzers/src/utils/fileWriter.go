package utils

import (
	"os"
	"path"

	"github.com/tony-507/analyzers/src/common"
)

type FileWriter interface {
	Open() error
	Write(common.CmBuf)
	Close() error
}

type CsvWriterStruct struct {
	fHandle *os.File
	fname   string
	hasHead bool
	outDir  string
}

func (csv *CsvWriterStruct) Open() error {
	fHandle, err := os.OpenFile(path.Join(csv.outDir, csv.fname), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	csv.fHandle = fHandle
	return nil
}

func (csv *CsvWriterStruct) Write(cmBuf common.CmBuf) {
	if (csv.fHandle == nil) {
		return
	}
	if !csv.hasHead {
		csv.fHandle.WriteString(cmBuf.GetFieldAsString())
		csv.hasHead = true
	}
	csv.fHandle.WriteString(cmBuf.ToString())
}

func (csv *CsvWriterStruct) Close() error {
	return csv.fHandle.Close()
}

func CsvWriter(outDir string, fname string) FileWriter {
	if _, dirErr := os.Stat(outDir); dirErr == nil {
		rv := &CsvWriterStruct{
			fHandle: nil,
			fname: fname,
			hasHead: false,
			outDir: outDir,
		}
		return rv
	}
	return nil
}


type RawWriterStruct struct {
	fHandle *os.File
	fname   string
	outDir  string
}

func (raw *RawWriterStruct) Open() error {
	fHandle, err := os.OpenFile(path.Join(raw.outDir, raw.fname), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	raw.fHandle = fHandle
	return nil
}

func (raw *RawWriterStruct) Write(cmBuf common.CmBuf) {
	if (raw.fHandle == nil) {
		return
	}
	raw.fHandle.Write(cmBuf.GetBuf())
}

func (raw *RawWriterStruct) Close() error {
	return raw.fHandle.Close()
}

func RawWriter(outDir string, fname string) FileWriter {
	if _, dirErr := os.Stat(outDir); dirErr == nil {
		rv := &RawWriterStruct{
			fHandle: nil,
			fname: fname,
			outDir: outDir,
		}
		return rv
	}
	return nil
}
