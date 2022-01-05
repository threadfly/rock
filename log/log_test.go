package log_test

import (
	"testing"

	"rock/log"
)

func TestLog(t *testing.T) {
	log.Debug("default log level, 1 log")
	log.Error("default log level, 2 log")
	log.Info("default log level, 3 log")
	log.SetLogLevel(log.ERROR)
	log.Debug("error log level, 1 log")
	log.Error("error log level, 2 log")
	log.Info("error log level, 3 log")
	log.SetLogLevel(log.INFO)
	log.Debug("info log level, 1 log")
	log.Error("info log level, 2 log")
	log.Info("info log level, 3 log")

	log.Debug("info log level, 1 log format")
	log.Error("info log level, 2 log format")
	log.Info("info log level, 3 log format")
	log.SetLogLevel(log.SemanticSwitch("Debug"))
	log.Debug("debug log level, 1 log format")
	log.SetLogLevel(log.SemanticSwitch("Debugf"))
}
