package api

import (
	"github.com/viliproject/vili/log"
	"github.com/viliproject/vili/slack"
)

func logMessage(message, slackMessage string, level log.Level) {
	var logf func(...interface{})
	switch level {
	case log.DebugLevel:
		logf = log.Debug
	case log.InfoLevel:
		logf = log.Info
	case log.WarnLevel:
		logf = log.Warn
	case log.ErrorLevel:
		logf = log.Error
	default:
		log.Errorf("Invalid level for logging message %v", level)
		return
	}
	logf(message)

	if level < log.DebugLevel {
		if level == log.ErrorLevel {
			slackMessage += " <!channel>"
		}
		err := slack.PostLogMessage(slackMessage, level)
		if err != nil {
			log.WithError(err).Error("Failed posting slack message")
		}
	}
}
