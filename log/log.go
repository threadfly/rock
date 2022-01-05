package log

type logAdaptor interface {
	Debug(...interface{})
	Debugf(format string, v ...interface{})
	Error(...interface{})
	Errorf(format string, v ...interface{})
	Info(...interface{})
	Infof(format string, v ...interface{})
}

type logLevel int8

const (
	DEBUG logLevel = iota
	ERROR
	INFO
)

func SemanticSwitch(lvl string) logLevel {
	switch lvl {
	case "error", "Error", "ERROR":
		return ERROR
	case "info", "Info", "INFO":
		return INFO
	case "debug", "Debug", "DEBUG":
	}
	return DEBUG
}

var (
	logLvl logLevel = DEBUG
	log    logAdaptor
)

func init() {
	log = Std(LONG_FILE_LOG_TYPE, 3)
}

func SetLogLevel(lvl logLevel) {
	logLvl = lvl
}

func SetLogAdaptor(adaptor logAdaptor) {
	log = adaptor
}

func Debug(v ...interface{}) {
	if logLvl <= DEBUG {
		log.Debug(v...)
	}
}

func Debugf(format string, v ...interface{}) {
	if logLvl <= DEBUG {
		log.Debugf(format, v...)
	}
}

func Error(v ...interface{}) {
	if logLvl <= ERROR {
		log.Error(v...)
	}
}

func Errorf(format string, v ...interface{}) {
	if logLvl <= ERROR {
		log.Errorf(format, v...)
	}
}

func Info(v ...interface{}) {
	if logLvl <= INFO {
		log.Info(v...)
	}
}

func Infof(format string, v ...interface{}) {
	if logLvl <= INFO {
		log.Infof(format, v...)
	}
}
