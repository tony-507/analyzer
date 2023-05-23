package common

import (
	"fmt"
	"log"
	"time"
)

// Logger struct
type Log struct {
	id string // Should be name of struct
}

func CreateLogger(id string) Log {
	return Log{id: id}
}

func (l *Log) Trace(msg string, param ...interface{}) {
	l._log(_LOG_TRACE, msg, param...)
}

func (l *Log) Info(msg string, param ...interface{}) {
	l._log(_LOG_INFO, msg, param...)
}

func (l *Log) Warn(msg string, param ...interface{}) {
	l._log(_LOG_WARN, msg, param...)
}

func (l *Log) Error(msg string, param ...interface{}) {
	l._log(_LOG_ERROR, msg, param...)
}

func (l *Log) Critical(msg string, param ...interface{}) {
	l._log(_LOG_CRITICAL, msg, param...)
}

func (l *Log) Fatal(msg string, param ...interface{}) {
	l._log(_LOG_FATAL, msg, param...)
}

func (l *Log) _log(level int, msg string, param ...interface{}) {
	// Check if user has setting
	if globalConfig.logLevel == 0 {
		globalConfig.logLevel = _LOG_TRACE
	}
	if globalConfig.msgPrefix == "" {
		globalConfig.msgPrefix = "[%l]"
	}

	if globalConfig.logLevel == _LOG_DISABLED {
		return
	}

	if level >= globalConfig.logLevel {
		sb := "[" + l.id + "] "

		// Use a string builder pattern to build the message
		bNextIsOpt := false
		for _, chr := range globalConfig.msgPrefix {
			// Start of an option
			if chr == '%' {
				bNextIsOpt = true
				continue
			}
			if bNextIsOpt {
				sb += l.getLogOpt(chr, level)
				bNextIsOpt = false
			} else {
				sb += string(chr)
			}
		}

		log.Printf(sb+" "+msg, param...)
	}
}

func (l *Log) getLogOpt(opt rune, level int) string {
	switch opt {
	case 't':
		return time.Now().UTC().String()
	case 'n':
		return l.id
	case 'l':
		switch level {
		case _LOG_TRACE:
			return "TRACE"
		case _LOG_INFO:
			return "INFO"
		case _LOG_WARN:
			return "WARN"
		case _LOG_ERROR:
			return "ERROR"
		case _LOG_CRITICAL:
			return "CRITICAL"
		case _LOG_FATAL:
			return "FATAL"
		default:
			fmt.Println("Unknown level:", level)
			panic("Logger internal error")
		}
	default:
		fmt.Println("Unknown option:", string(opt))
		panic("Logger internal error")
	}
}
