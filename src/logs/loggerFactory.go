package logs

const (
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
	msgFormat string
}

var globalConfig globalLoggerConfig

// Setting a property for the global logger
func SetProperty(key string, val string) {
	switch key {
	case "level":
		switch val {
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
	case "format":
		globalConfig.msgFormat = val
	default:
		panic("Unknown property for Logger")
	}
}