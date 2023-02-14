package common

import (
	"log"
	"os"
)

const (
	_LOG_DISABLED int = 0
	_LOG_TRACE    int = 1
	_LOG_INFO     int = 2
	_LOG_WARN     int = 3
	_LOG_ERROR    int = 4
	_LOG_CRITICAL int = 5
	_LOG_FATAL    int = 6
)

// Global configuration for logger
type loggerConfig struct {
	logLevel  int
	msgPrefix string
	logFile   *os.File
}

var globalConfig loggerConfig

// Setting a property for the global logger
func SetLoggingProperty(key string, val string) {
	switch key {
	case "level":
		switch val {
		case "disabled":
			globalConfig.logLevel = _LOG_DISABLED
		case "trace":
			globalConfig.logLevel = _LOG_TRACE
		case "info":
			globalConfig.logLevel = _LOG_INFO
		case "warn":
			globalConfig.logLevel = _LOG_WARN
		case "critical":
			globalConfig.logLevel = _LOG_CRITICAL
		case "fatal":
			globalConfig.logLevel = _LOG_FATAL
		default:
			panic("Error in setting logger level")
		}
	case "prefix":
		globalConfig.msgPrefix = val
	case "logDir":
		os.MkdirAll(val, os.ModePerm)
		f, err := os.Create(val + "/app.log")
		if err != nil {
			panic(err)
		}
		globalConfig.logFile = f
		log.SetOutput(f)
	default:
		panic("Unknown property for Logger")
	}
}

func CleanUp() {
	if globalConfig.logFile != nil {
		globalConfig.logFile.Close()
	}
}
