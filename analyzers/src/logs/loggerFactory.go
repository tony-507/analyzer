package logs

import (
	"log"
	"os"
)

const (
	DISABLED int = 0
	TRACE    int = 1
	INFO     int = 2
	WARN     int = 3
	ERROR    int = 4
	CRITICAL int = 5
	FATAL    int = 6
)

// Global configuration for logger
type globalLoggerConfig struct {
	logLevel  int
	msgPrefix string
	logFile   *os.File
}

var globalConfig globalLoggerConfig

// Setting a property for the global logger
func SetProperty(key string, val string) {
	switch key {
	case "level":
		switch val {
		case "disabled":
			globalConfig.logLevel = DISABLED
		case "trace":
			globalConfig.logLevel = TRACE
		case "info":
			globalConfig.logLevel = INFO
		case "warn":
			globalConfig.logLevel = WARN
		case "critical":
			globalConfig.logLevel = CRITICAL
		case "fatal":
			globalConfig.logLevel = FATAL
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
