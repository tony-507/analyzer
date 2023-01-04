package logs

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

func (l *Log) Log(level int, msg string, param ...interface{}) {
	// Check if user has setting
	if globalConfig.logLevel == 0 {
		globalConfig.logLevel = TRACE
	}
	if globalConfig.msgPrefix == "" {
		globalConfig.msgPrefix = "[%l]"
	}

	if globalConfig.logLevel == DISABLED {
		return
	}

	if level >= globalConfig.logLevel {
		sb := ""

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
		case TRACE:
			return "TRACE"
		case INFO:
			return "INFO"
		case WARN:
			return "WARN"
		case ERROR:
			return "ERROR"
		case CRITICAL:
			return "CRITICAL"
		case FATAL:
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
