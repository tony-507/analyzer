package ioUtils

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	tm "github.com/buger/goterm"
	"github.com/tony-507/analyzers/src/common"
)

var reader = bufio.NewReader(os.Stdin)

var Reset = "\033[0m"
var Red = "\033[31m"
var Green = "\033[32m"
var Yellow = "\033[33m"
var Blue = "\033[34m"
var Purple = "\033[35m"
var Cyan = "\033[36m"
var Gray = "\033[37m"
var White = "\033[97m"

type row struct {
	pktCnt          int
	pts             int
	auPerPes        int
	spliceCountdown int
}

type queue struct {
	pid           int
	reserveBuffer int // Reserve first n records
	rows          []row
}

func (q *queue) drop(idx int) {
	if idx == 0 {
		q.rows = q.rows[1:]
	} else {
		q.rows = append(q.rows[:idx], q.rows[(idx+1):]...)
	}
}

func (q *queue) append(r row, maxSize int, atEnd bool) {
	if len(q.rows) > maxSize {
		q.drop(q.reserveBuffer)
	}
	if atEnd {
		q.rows = append(q.rows, r)
	} else {
		if q.reserveBuffer > len(q.rows) {
			q.rows = append(q.rows, r)
		} else {
			q.rows[q.reserveBuffer-1] = r
		}
	}
}

type adSmartMonitor struct {
	logger        common.Log
	isRunning     bool
	bufferSize    int
	nextSplicePTS int
	videoQueue    queue
	audioQueue    queue
	dataQueue     queue
}

func (w *adSmartMonitor) setup(param ioWriterParam) {
	w.bufferSize = 6
	w.isRunning = true
	w.videoQueue = queue{rows: make([]row, 0), pid: 32}
	w.audioQueue = queue{rows: make([]row, 0), pid: 33}
	w.dataQueue = queue{rows: make([]row, 0), pid: 35}
	go w.monitorStreams()
}

func (w *adSmartMonitor) stop() {
	w.isRunning = false
}

func (w *adSmartMonitor) processControl(unit common.CmUnit) {}

func (w *adSmartMonitor) processUnit(unit common.CmUnit) {
	if unit == nil {
		return
	}

	pid, isInt := unit.GetField("id").(int)
	if pid == 0 || pid == 480 {
		return
	}
	if !isInt {
		panic("id is not pid")
	}

	if cmBuf, isCmBuf := unit.GetBuf().(common.CmBuf); isCmBuf {
		pktCnt := -1
		pts := -1
		auPerPes := -1
		spliceCountdown := 100 // Special hard-coded value
		appendAtEnd := true

		f, ok := cmBuf.GetField("pktCnt")
		if !ok {
			panic("No pktCnt")
		}
		pktCnt, ok = f.(int)
		if !ok {
			panic("pktCnt not int")
		}

		f, ok = cmBuf.GetField("PTS")
		if !ok {
			panic("No PTS")
		}
		pts, ok = f.(int)
		if !ok {
			panic("PTS not int")
		}

		if pid == 35 {
			w.logger.Info("PktCnt: %d, PTS: %d", pktCnt, pts)
		}

		// Audio only
		if pid == 33 {
			f, ok := cmBuf.GetField("auPerPes")
			if !ok {
				panic("No auPerPes")
			}
			auPerPes, ok = f.(int)
			if !ok {
				panic("AuPerPes not int")
			}
		}

		// Video only
		if pid == 32 {
			f, ok := cmBuf.GetField("spliceCountdown")
			if !ok {
				panic("No spliceCountdown")
			}
			spliceCountdown, ok = f.(int)
			if !ok {
				panic("spliceCountdown not int")
			}
		}

		switch pid {
		case 32:
			if spliceCountdown != 100 {
				w.logger.Info("Received splice_countdown %d", spliceCountdown)
				w.videoQueue.reserveBuffer += 1
				appendAtEnd = false
			}
			w.videoQueue.append(row{pktCnt: pktCnt, pts: pts, auPerPes: auPerPes, spliceCountdown: spliceCountdown}, w.bufferSize, appendAtEnd)
		case 33:
			if auPerPes != 6 {
				w.logger.Info("Received special AU structure: %d", auPerPes)
				w.audioQueue.reserveBuffer += 1
				appendAtEnd = false
			}
			w.audioQueue.append(row{pktCnt: pktCnt, pts: pts, auPerPes: auPerPes, spliceCountdown: spliceCountdown}, w.bufferSize, appendAtEnd)
		case 35:
			if pts != -1 {
				w.logger.Info("Receive SCTE-35")
				w.nextSplicePTS = pts
				w.dataQueue.reserveBuffer += 1
				appendAtEnd = false
				w.dataQueue.append(row{pktCnt: pktCnt, pts: pts, auPerPes: auPerPes, spliceCountdown: spliceCountdown}, w.bufferSize, appendAtEnd)
			}
		}

	} else {
		panic(fmt.Sprintf("What is this? %T | %v", unit.GetBuf(), unit.GetBuf()))
	}
}

func (q *queue) getOutput(bufferSize int, nextSplicePTS int) string {
	rows := q.rows
	msg := ""
	l := len(rows)
	for i := 0; i < bufferSize; i++ {
		pidHeading := ""
		heading := ""
		if i == 0 {
			switch q.pid {
			case 32:
				pidHeading = "32"
				heading += "H.264"
			case 33:
				pidHeading = "33"
				heading += "E-AC-3"
			case 35:
				pidHeading = "35"
				heading += "SCTE-35"
			}
		}
		pktCntStr := strings.Repeat(" ", 15)
		ptsStr := strings.Repeat(" ", 10)
		auStr := strings.Repeat(" ", 10)
		spliceCountdownStr := strings.Repeat(" ", 15)
		noteStr := strings.Repeat(" ", 35)
		colorStr := Reset
		if i < l {
			if rows[i].pktCnt != -1 {
				pktCntStr = Reset + fmt.Sprintf("%15d", rows[i].pktCnt)
			}
			if rows[i].pts != -1 {
				if q.pid == 35 {
					colorStr = Yellow
				} else if q.pid == 32 && rows[i].pts == nextSplicePTS {
					colorStr = Yellow
				}
				ptsStr = colorStr + fmt.Sprintf("%10d", rows[i].pts)
			}
			if rows[i].auPerPes != -1 {
				if rows[i].auPerPes == 1 {
					if rows[i].pts < nextSplicePTS {
						colorStr = Red
					} else {
						colorStr = Green
					}
				}
				auStr = colorStr + fmt.Sprintf("%-15d", rows[i].auPerPes)
				noteStr = fmt.Sprintf("%15s  =  %s", "auPerPes", auStr)
			}
			if rows[i].spliceCountdown != 100 {
				colorStr = Yellow
				spliceCountdownStr = colorStr + fmt.Sprintf("%-15d", rows[i].spliceCountdown)
				noteStr = fmt.Sprintf("%15s  =  %s", "SpliceCountdown", spliceCountdownStr)
			}
		}
		msg += fmt.Sprintf(Reset+"| %5s"+Reset+" | %10s"+Reset+" | "+"%s"+Reset+" | "+"%s"+Reset+" | "+"%s"+Reset+" |\n",
			pidHeading, heading, pktCntStr, ptsStr, noteStr)
	}
	q.reserveBuffer = 0
	return msg
}

func (w *adSmartMonitor) monitorStreams() {
	normalSleepDuration := 3 * time.Second
	demoSleepDuration := 20 * time.Second
	sepStr := strings.Repeat("-", 89)
	for w.isRunning {
		tm.Clear()
		tm.MoveCursor(1, 1)
		demoInProgress := (w.videoQueue.reserveBuffer != 0) || (w.audioQueue.reserveBuffer != 0)
		msg := ""
		msg += fmt.Sprintf(Reset+"-%s-\n", sepStr)
		msg += fmt.Sprintf(Reset+"| %5s | %10s | %15s | %10s | %35s |\n", "Pid", "Stream", "Packet index", "PTS", "Notes")
		msg += fmt.Sprintf(Reset+"|%s|\n", sepStr)
		msg += w.videoQueue.getOutput(w.bufferSize, w.nextSplicePTS)
		msg += fmt.Sprintf(Reset+"|%s|\n", sepStr)
		msg += w.audioQueue.getOutput(w.bufferSize, w.nextSplicePTS)
		msg += fmt.Sprintf(Reset+"|%s|\n", sepStr)
		msg += w.dataQueue.getOutput(w.bufferSize, w.nextSplicePTS)
		msg += fmt.Sprintf(Reset+"-%s-\n", sepStr)
		tm.Println(msg)
		tm.Flush()
		if demoInProgress {
			time.Sleep(demoSleepDuration)
		} else {
			time.Sleep(normalSleepDuration)
		}
	}
}

func getAdSmartMonitor(name string) *adSmartMonitor {
	return &adSmartMonitor{logger: common.CreateLogger(name), isRunning: false, bufferSize: 0}
}
