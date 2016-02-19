package log

import (
	"github.com/Sirupsen/logrus"
)

var logger = logrus.New()

// Config is the configuration for the odin logger
type Config struct {
	LogJSON      bool
	LogDebug     bool
	LongFileName bool
}

// Init initializes the odin logger with the given config
func Init(config *Config) {
	if config.LogJSON {
		logger.Formatter = &logrus.JSONFormatter{}
	} else {
		logger.Formatter = &logrus.TextFormatter{FullTimestamp: true}
	}
	if config.LogDebug {
		logger.Level = logrus.DebugLevel
	}
	shortenFilename = !config.LongFileName
}

// GetLogger returns the logger
func GetLogger() *logrus.Logger {
	return logger
}
