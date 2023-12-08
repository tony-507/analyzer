package fileReader

import (
	"errors"
	"fmt"
	"os"
)

type binaryData struct {
	fHandle *os.File
	hasData bool
}

func (bf *binaryData) getBuffer() ([]byte, error) {
	if !bf.hasData {
		return []byte{}, nil
	}
	stat, err := bf.fHandle.Stat()
	if err != nil {
		return []byte{}, errors.New(fmt.Sprintf("Fail to retrieve file stat: %s", err.Error()))
	}
	buf := make([]byte, stat.Size())
	_, err = bf.fHandle.Read(buf)
	if err != nil {
		return buf, errors.New(fmt.Sprintf("Fail to read buffer: %s", err.Error()))
	}
	bf.hasData = false
	return buf, nil
}

func (bf *binaryData) tick() (int64, bool) {
	return 0, false
}

func binaryDataFile(fHandle *os.File) fileHandler {
	return &binaryData{
		fHandle: fHandle,
		hasData: true,
	}
}
