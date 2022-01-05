package log

type LogAdaptor struct {
	logLvl logLevel
	log    logAdaptor
}

func NewLogAdaptor(adp logAdaptor) *LogAdaptor {
	if adp == nil {
		adp = Std(SHORT_FILE_LOG_TYPE, 3)
	}

	return &LogAdaptor{
		logLvl: DEBUG,
		log:    adp,
	}
}

func (l *LogAdaptor) SetLogAdaptor(adaptor logAdaptor) {
	l.log = adaptor
}

func (l *LogAdaptor) SetLogLevel(lvl string) {
	l.logLvl = SemanticSwitch(lvl)
}

func (l *LogAdaptor) Debug(v ...interface{}) {
	if l.logLvl <= DEBUG {
		l.log.Debug(v...)
	}
}

func (l *LogAdaptor) Debugf(format string, v ...interface{}) {
	if l.logLvl <= DEBUG {
		l.log.Debugf(format, v...)
	}
}

func (l *LogAdaptor) Error(v ...interface{}) {
	if l.logLvl <= ERROR {
		l.log.Error(v...)
	}
}

func (l *LogAdaptor) Errorf(format string, v ...interface{}) {
	if l.logLvl <= ERROR {
		l.log.Errorf(format, v...)
	}
}

func (l *LogAdaptor) Info(v ...interface{}) {
	if l.logLvl <= INFO {
		l.log.Info(v...)
	}
}

func (l *LogAdaptor) Infof(format string, v ...interface{}) {
	if l.logLvl <= INFO {
		l.log.Infof(format, v...)
	}
}
