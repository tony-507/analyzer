package ioUtils

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"sync"

	"github.com/tony-507/analyzers/src/common"
)

/*
FileWriter

An object to write buffer into a file
- Holds multiple file handles in a folder
- Writes various types of outputs, e.g. json, csv
*/

type FileWriter struct {
	logger     common.Log
	writerMap  []chan common.CmUnit // Pre-assign a fixed number of channels to prevent race condition during runtime channel creation
	rawByteExt string
	idMapping  []string // This maps id to channel index
	outFolder  string
	wg         sync.WaitGroup
}

func (m_writer *FileWriter) setup(writerParam ioWriterParam) error {
	m_writer.writerMap = make([]chan common.CmUnit, 40)
	m_writer.outFolder = writerParam.FileOutput.OutFolder
	m_writer.rawByteExt = writerParam.FileOutput.RawByteExtension
	for i := range m_writer.writerMap {
		m_writer.writerMap[i] = make(chan common.CmUnit)
	}

	return os.MkdirAll(m_writer.outFolder, os.ModePerm) // Create output folder if necessary
}

func (m_writer *FileWriter) stop() {
	for idx := range m_writer.idMapping {
		stopUnit := common.MakeStatusUnit(common.STATUS_END_ROUTINE, nil)
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
		if id == strconv.Itoa(outId) {
			idIdx = idx
		}
	}

	if idIdx == -1 {
		m_writer.logger.Error("Handler not found for unit: %v", unit)
	}

	m_writer.writerMap[idIdx] <- unit
}

func (m_writer *FileWriter) _processJsonOutput(id string, chIdx int) {
	defer m_writer.wg.Done()
	m_writer.logger.Trace("JSON handler with id %s starts", id)

	isInit := false

	fname := fmt.Sprintf("%s%s.json", m_writer.outFolder, id)
	f, err := os.Create(fname)
	if err != nil {
		m_writer.logger.Error("Fail to create and open %s: %s", fname, err.Error())
	}

	defer f.Close()

	// We are writing array of jsons, so add first line
	f.Write([]byte("[\n"))

	for {
		unit := <-m_writer.writerMap[chIdx]
		_, isStatus := unit.(*common.CmStatusUnit)

		if isStatus {
			break
		}

		if cmBuf, isCmBuf := unit.GetBuf().(common.CmBuf); isCmBuf {
			buf := cmBuf.GetBuf()

			if len(buf) == 1 {
				break
			} else if isInit {
				f.Write([]byte(","))
			} else {
				f.Write([]byte("\t"))
			}

			_, err := f.Write(buf)
			if err != nil {
				m_writer.logger.Error("Error writing data (%v) to %s: %s", buf, fname, err.Error())
			}
		} else {
			m_writer.logger.Error("Received unknown unit buffer: %v", unit.GetBuf())
		}
	}

	f.Write([]byte("\n]"))
	m_writer.logger.Trace("JSON handler with id %s stops", id)
}

func (m_writer *FileWriter) _processCsvOutput(id string, chIdx int) {
	defer m_writer.wg.Done()
	m_writer.logger.Trace("CSV handler with id %s starts", id)

	fname := ""
	rawFileName := ""
	if true {
		fname = fmt.Sprintf("%s%s.csv", m_writer.outFolder, id)
		rawFileName = fmt.Sprintf("%sraw_%s.%s", m_writer.outFolder, id, m_writer.rawByteExt)
	} else {
		fname = fmt.Sprintf("%s%s.csv", m_writer.outFolder, "packets")
		rawFileName = fmt.Sprintf("%sraw_%s.%s", m_writer.outFolder, "packets", m_writer.rawByteExt)
	}

	b_HasHeader := false // Header has been written
	header := ""
	body := ""

	csvFile, err := os.Create(fname)
	if err != nil {
		m_writer.logger.Error("Fail to create and open %s: %s", fname, err.Error())
	}
	csvWriter := bufio.NewWriter(csvFile)

	rawFile, err := os.Create(rawFileName)
	if err != nil {
		m_writer.logger.Error("Fail to create and open %s: %s", rawFileName, err.Error())
	}

	defer csvFile.Close()

	for {
		unit := <-m_writer.writerMap[chIdx]
		_, isStatus := unit.(*common.CmStatusUnit)

		if isStatus {
			break
		}

		if cmBuf, isCmBuf := unit.GetBuf().(common.CmBuf); isCmBuf {
			header = cmBuf.GetFieldAsString()
			body = cmBuf.ToString()
			data := cmBuf.GetBuf()
			rawFile.Write(data)
		}

		if !b_HasHeader {
			csvWriter.WriteString(header)
			b_HasHeader = true
		}

		_, err := csvWriter.WriteString(body)
		if err != nil {
			m_writer.logger.Error("Error writing data (%s) to %s: %s", body, fname, err.Error())
		}

		csvWriter.Flush()
	}
	m_writer.logger.Trace("CSV handler with id %s stops", id)
}

func (m_writer *FileWriter) _processRawOutput(id string, chIdx int) {
	defer m_writer.wg.Done()
	m_writer.logger.Trace("Raw handler with id %s starts", id)

	fname := fmt.Sprintf("%sout%s.ts", m_writer.outFolder, id)
	f, err := os.Create(fname)
	if err != nil {
		m_writer.logger.Error("Fail to create and open %s: %s", fname, err.Error())
	}

	for {
		unit := <-m_writer.writerMap[chIdx]
		_, isStatus := unit.(*common.CmStatusUnit)

		if isStatus {
			break
		}

		buf, _ := unit.GetBuf().([]byte)

		_, err := f.Write(buf)
		if err != nil {
			m_writer.logger.Error("Error writing data (%v) to %s: %s", buf, fname, err.Error())
		}
	}
	m_writer.logger.Trace("Raw handler with id %s stops", id)
}

func (m_writer *FileWriter) processControl(unit common.CmUnit) {
	id, isValidId := unit.GetField("id").(int)
	if !isValidId {
		return
	}
	buf, isValidBuf := unit.GetBuf().(common.CmBuf)
	if !isValidBuf {
		return
	}
	if id == 0x10 {
		field, hasField := buf.GetField("addId")
		if !hasField {
			return
		}
		addId, isBool := field.(bool)
		if !isBool {
			return
		}

		field, hasField = buf.GetField("id")
		if !hasField {
			return
		}
		outId, isString := field.(string)
		if !isString {
			return
		}

		field, hasField = buf.GetField("type")
		outType, isInt := field.(int)
		if !isInt {
			return
		}

		if addId {
			idIdx := len(m_writer.idMapping)
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
				m_writer.logger.Error("Received unknown output type %d for id %d", outType, idIdx)
			}
		}
	}
}

func getFileWriter(name string) *FileWriter {
	rv := FileWriter{logger: common.CreateLogger(name)}
	return &rv
}
