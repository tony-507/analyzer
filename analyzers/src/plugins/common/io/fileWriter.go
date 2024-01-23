package io

import (
	"os"
	"path"

	"github.com/tony-507/analyzers/src/tttKernel"
)

type FileWriter interface {
	Open() error
	Write(tttKernel.CmBuf)
	Close() error
}

type CsvWriterStruct struct {
	fHandle *os.File
	fname   string
	hasHead bool
	outDir  string
}

func (csv *CsvWriterStruct) Open() error {
	if _, dirErr := os.Stat(csv.outDir); dirErr != nil {
		return dirErr
	}
	fHandle, err := os.OpenFile(path.Join(csv.outDir, csv.fname), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	csv.fHandle = fHandle
	return nil
}

func (csv *CsvWriterStruct) Write(cmBuf tttKernel.CmBuf) {
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
	if csv.fHandle == nil {
		return nil
	}
	return csv.fHandle.Close()
}

func CsvWriter(outDir string, fname string) FileWriter {
	return &CsvWriterStruct{
		fHandle: nil,
		fname: fname,
		hasHead: false,
		outDir: outDir,
	}
}


type RawWriterStruct struct {
	fHandle *os.File
	fname   string
	outDir  string
}

func (raw *RawWriterStruct) Open() error {
	if _, dirErr := os.Stat(raw.outDir); dirErr != nil {
		return dirErr
	}
	fHandle, err := os.OpenFile(path.Join(raw.outDir, raw.fname), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	raw.fHandle = fHandle
	return nil
}

func (raw *RawWriterStruct) Write(cmBuf tttKernel.CmBuf) {
	if (raw.fHandle == nil) {
		return
	}
	raw.fHandle.Write(cmBuf.GetBuf())
}

func (raw *RawWriterStruct) Close() error {
	if raw.fHandle == nil {
		return nil
	}
	return raw.fHandle.Close()
}

func RawWriter(outDir string, fname string) FileWriter {
	return &RawWriterStruct{
		fHandle: nil,
		fname: fname,
		outDir: outDir,
	}
}
