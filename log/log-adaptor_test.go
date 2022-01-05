package log_test

import (
	"testing"

	"rock/log"
)

func TestLogAdaptor(t *testing.T) {
	l := log.NewLogAdaptor(nil)
	l.Debug("default log level, 1 log")
	l.Error("default log level, 2 log")
	l.Info("default log level, 3 log")
	l.SetLogLevel("error")
	l.Debug("error log level, 1 log")
	l.Error("error log level, 2 log")
	l.Info("error log level, 3 log")
	l.SetLogLevel("Info")
	l.Debug("info log level, 1 log")
	l.Error("info log level, 2 log")
	l.Info("info log level, 3 log")

	l.SetLogAdaptor(log.Std(log.LONG_FILE_LOG_TYPE, 3))
	l.Debug("info log level, 1 log format")
	l.Error("info log level, 2 log format")
	l.Info("info log level, 3 log format")
	l.SetLogLevel("Debug")
	l.Debug("debug log level, 1 log format")
	l.SetLogLevel("Debugf")
}
