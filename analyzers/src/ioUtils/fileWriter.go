package ioUtils

import (
	"bufio"
	"fmt"
	"os"
	"sync"

	"github.com/tony-507/analyzers/src/common"
	"github.com/tony-507/analyzers/src/logs"
)

/*
FileWriter

An object to write buffer into a file
- Holds multiple file handles in a folder
- Writes various types of outputs, e.g. json, csv
*/

type FileWriter struct {
	logger    logs.Log
	writerMap []chan common.CmUnit // Pre-assign a fixed number of channels to prevent race condition during runtime channel creation
	idMapping []int                // This maps id to channel index
	outFolder string
	wg        sync.WaitGroup
}

func (m_writer *FileWriter) setup(writerParam IOWriterParam) {
	m_writer.logger = logs.CreateLogger("FileWriter")
	m_writer.writerMap = make([]chan common.CmUnit, 40)
	m_writer.outFolder = writerParam.FileOutput.OutFolder
	for i := range m_writer.writerMap {
		m_writer.writerMap[i] = make(chan common.CmUnit)
	}

	err := os.MkdirAll(m_writer.outFolder, os.ModePerm) // Create output folder if necessary
	if err != nil {
		panic(err)
	}
}

func (m_writer *FileWriter) stop() {
	for idx := range m_writer.idMapping {
		stopUnit := common.MakeStatusUnit(common.STATUS_END_ROUTINE, 1, "")
		m_writer.writerMap[idx] <- stopUnit
	}

	m_writer.wg.Wait()
}

func (m_writer *FileWriter) processUnit(unit common.CmUnit) {
	if unit == nil {
		return
	}

	outId, _ := unit.GetField("id").(int)
	if outId == 1 {
		return
	}
	idIdx := -1
	for idx, id := range m_writer.idMapping {
		if id == outId {
			idIdx = idx
		}
	}

	if idIdx == -1 {
		idIdx = len(m_writer.idMapping)

		outType := unit.GetField("type")
		m_writer.idMapping = append(m_writer.idMapping, outId)

		switch outType {
		case 0:
			// Undefined, skip
		case 1:
			m_writer.wg.Add(1)
			go m_writer._processCsvOutput(outId, idIdx)
		case 2:
			m_writer.wg.Add(1)
			go m_writer._processJsonOutput(outId, idIdx)
		case 3:
			m_writer.wg.Add(1)
			go m_writer._processRawOutput(outId, idIdx)
		default:
			m_writer.logger.Log(logs.ERROR, "unknown output type %v for id %v", outType, idIdx)
			panic("Unknown output type")
		}
	}
	m_writer.writerMap[idIdx] <- unit
}

func (m_writer *FileWriter) _processJsonOutput(pid int, chIdx int) {
	defer m_writer.wg.Done()
	m_writer.logger.Log(logs.TRACE, "JSON handler for pid ", pid, " starts")

	isInit := false

	fname := fmt.Sprintf("%s%d.json", m_writer.outFolder, pid)
	f, err := os.Create(fname)
	check(err)

	defer f.Close()

	// We are writing array of jsons, so add first line
	f.Write([]byte("{[\n"))

	for {
		unit := <-m_writer.writerMap[chIdx]
		_, isStatus := unit.(common.CmStatusUnit)

		if isStatus {
			break
		}

		buf, _ := unit.GetBuf().([]byte)

		if len(buf) == 1 {
			break
		} else if isInit {
			f.Write([]byte(","))
		} else {
			f.Write([]byte("\t"))
		}

		_, err := f.Write(buf)
		check(err)
	}

	f.Write([]byte("\n]}"))
	m_writer.logger.Log(logs.TRACE, "JSON handler for pid ", pid, " stops")
}

func (m_writer *FileWriter) _processCsvOutput(pid int, chIdx int) {
	defer m_writer.wg.Done()
	m_writer.logger.Log(logs.TRACE, "CSV handler for pid ", pid, " starts")

	fname := ""
	if pid != -1 {
		fname = fmt.Sprintf("%s%d.csv", m_writer.outFolder, pid)
	} else {
		fname = fmt.Sprintf("%s%s.csv", m_writer.outFolder, "packets")
	}

	b_HasHeader := false // Header has been written
	header := ""
	body := ""

	f, err := os.Create(fname)
	w := bufio.NewWriter(f)
	check(err)

	defer f.Close()

	for {
		unit := <-m_writer.writerMap[chIdx]
		_, isStatus := unit.(common.CmStatusUnit)

		if isStatus {
			break
		}

		if cmBuf, isCmBuf := unit.GetBuf().(common.SimpleBuf); isCmBuf {
			header = cmBuf.GetFieldAsString()
			body = cmBuf.ToString()
		}

		if !b_HasHeader {
			w.WriteString(header)
			b_HasHeader = true
		}

		_, err := w.WriteString(body)
		check(err)

		w.Flush()
	}
	m_writer.logger.Log(logs.TRACE, "CSV handler for pid ", pid, " stops")
}

func (m_writer *FileWriter) _processRawOutput(pid int, chIdx int) {
	defer m_writer.wg.Done()
	m_writer.logger.Log(logs.TRACE, "Raw handler for pid ", pid, " starts")

	fname := fmt.Sprintf("%sout.ts", m_writer.outFolder)
	f, err := os.Create(fname)
	check(err)

	for {
		unit := <-m_writer.writerMap[chIdx]
		_, isStatus := unit.(common.CmStatusUnit)

		if isStatus {
			break
		}

		buf, _ := unit.GetBuf().([]byte)

		_, err := f.Write(buf)
		check(err)
	}
	m_writer.logger.Log(logs.TRACE, "Raw handler for pid ", pid, " ends")
}

func GetFileWriter() *FileWriter {
	rv := FileWriter{logger: logs.CreateLogger("fileWriter")}
	return &rv
}
