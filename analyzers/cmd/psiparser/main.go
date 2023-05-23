package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/tony-507/analyzers/src/plugins/avContainer/model"
)

type psiCallback struct{}

func (m *psiCallback) AddStream(version int, progNum int, streamPid int, streamType int) {}

func (m *psiCallback) AddProgram(version int, progNum int, pmtPid int) {}

func (m *psiCallback) GetPATVersion() int {
	return -1
}

func (m *psiCallback) GetPmtVersion(progNUm int) int {
	return -1
}

func (m *psiCallback) GetPmtPidByProgNum(progNum int) int {
	return -1
}

func (m *psiCallback) PsiUpdateFinished(pid int, jsonBytes []byte) {
	fmt.Println(string(jsonBytes))
}

func (m *psiCallback) SpliceEventReceived(dpiPid int, spliceCmdTypeStr string, splicePTS []int) {}

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Invalid number of arguments")
		fmt.Println("Usage: psiparser <byte_string>")
		os.Exit(1)
	}

	manager := &psiCallback{}
	inputBytes := []byte{}
	for _, v := range strings.Split(os.Args[1], " ") {
		intVal, _ := strconv.ParseInt(v, 16, 0)
		inputBytes = append(inputBytes, byte(intVal))
	}

	ds, err := model.PsiTable(manager, 0, 0, inputBytes)
	if err != nil {
		panic(err)
	}

	fmt.Println(fmt.Sprintf("Type: %s", ds.GetName()))

	err = ds.Process()
	if err != nil {
		panic(err)
	}
}
