package ioUtils

import (
	"bufio"
	"fmt"
	"os"
	"sync"
	"time"

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
	logger       logs.Log
	writerMap    []chan common.CmUnit // Pre-assign a fixed number of channels to prevent race condition during runtime channel creation
	idMapping    []int                // This maps id to channel index
	outFolder    string
	processedCnt int // How many units are processed
	isRunning    bool
	name         string
	wg           sync.WaitGroup
	monitorMtx   sync.Mutex // A mutex for locking monitoring variables
}

func (m_writer *FileWriter) SetParameter(m_parameter interface{}) {
	writerParam, isWriterParam := m_parameter.(IOWriterParam)
	if !isWriterParam {
		panic("Writer param has unknown format")
	}
	m_writer.outFolder = writerParam.OutFolder
	m_writer._setup()
}

func (m_writer *FileWriter) _setup() {
	m_writer.logger = logs.CreateLogger("FileWriter")
	m_writer.writerMap = make([]chan common.CmUnit, 40)
	for i := range m_writer.writerMap {
		m_writer.writerMap[i] = make(chan common.CmUnit, 1)
	}
	m_writer.isRunning = true
	m_writer.processedCnt = 0

	err := os.MkdirAll(m_writer.outFolder, os.ModePerm) // Create output folder if necessary
	if err != nil {
		panic(err)
	}

	m_writer.wg.Add(1)
	go m_writer._setupMonitor()
}

func (m_writer *FileWriter) _setupMonitor() {
	defer m_writer.wg.Done()

	orginialCnt := 0 // Check if the writer is still processing units

	for {
		if !m_writer.isRunning {
			break
		}

		time.Sleep(10 * time.Second)

		m_writer.monitorMtx.Lock()
		if m_writer.processedCnt == orginialCnt {
			statMsg := "\nFile writer status"
			statMsg += fmt.Sprintf("isRunning: %v", m_writer.isRunning)
			statMsg += fmt.Sprintf("File handles: %v", m_writer.idMapping)
			m_writer.logger.Log(logs.INFO, statMsg)
		} else {
			orginialCnt = m_writer.processedCnt
		}
		m_writer.monitorMtx.Unlock()
	}
}

func (m_writer *FileWriter) StartSequence() {
	m_writer.logger.Log(logs.INFO, "File writer is started")
}

func (m_writer *FileWriter) EndSequence() {
	for _, writer := range m_writer.writerMap {
		stopUnit := common.MakeStatusUnit(common.STATUS_END_ROUTINE, 1, "")
		writer <- stopUnit
	}
	m_writer.isRunning = false

	m_writer.wg.Wait()
}

func (m_writer *FileWriter) DeliverUnit(unit common.CmUnit) common.CmUnit {
	if unit == nil {
		return nil
	}

	outId, _ := unit.GetField("id").(int)
	if outId == 1 {
		return nil
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
		case 1:
			m_writer.wg.Add(1)
			go m_writer._processCsvOutput(outId, idIdx)
		case 2:
			m_writer.wg.Add(1)
			go m_writer._processJsonOutput(outId, idIdx)
		default:
			m_writer.logger.Log(logs.ERROR, "unknown output type %v for id %v", outType, idIdx)
			panic("Unknown output type")
		}
	}
	m_writer.writerMap[idIdx] <- unit

	return nil
}

func (m_writer *FileWriter) _processJsonOutput(pid int, chIdx int) {
	defer m_writer.wg.Done()

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

		m_writer.monitorMtx.Lock()
		m_writer.processedCnt += 1
		m_writer.monitorMtx.Unlock()
	}

	f.Write([]byte("\n]}"))
}

func (m_writer *FileWriter) _processCsvOutput(pid int, chIdx int) {
	defer m_writer.wg.Done()

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

		pesBuf, isPesBuf := unit.GetBuf().(common.PesBuf)
		if isPesBuf {
			header = pesBuf.GetFieldAsString()
			body = pesBuf.ToString()
		} else {
			psiBuf, isPsiBuf := unit.GetBuf().(common.PsiBuf)
			if isPsiBuf {
				header = psiBuf.GetFieldAsString()
				body = psiBuf.ToString()
			}
		}

		if !b_HasHeader {
			w.WriteString(header)
			b_HasHeader = true
		}

		_, err := w.WriteString(body)
		check(err)

		w.Flush()

		m_writer.monitorMtx.Lock()
		m_writer.processedCnt += 1
		m_writer.monitorMtx.Unlock()
	}
}

func (m_writer *FileWriter) FetchUnit() common.CmUnit {
	return nil
}

func GetFileWriter(name string) FileWriter {
	rv := FileWriter{name: name, logger: logs.CreateLogger("fileWriter")}
	return rv
}
