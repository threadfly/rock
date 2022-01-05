package log

import (
	"fmt"
	gov "log"
	"os"
)

type std int32

const (
	DEPTH_MASK    = 0x0F
	LOG_TYPE_MASK = 0xF0
)

type LOG_TYPE int32

const (
	LONG_FILE_LOG_TYPE  LOG_TYPE = 0x00
	SHORT_FILE_LOG_TYPE LOG_TYPE = 0x10
)

var (
	govStdShortLog = gov.New(os.Stderr, "", gov.LstdFlags|gov.Lshortfile)
	govStdLongLog  = gov.New(os.Stderr, "", gov.LstdFlags|gov.Llongfile)
)

func Std(fileTyp LOG_TYPE, depth int32) std {
	return std(int32(fileTyp) | depth)
}

func govLog(s std) *gov.Logger {
	switch LOG_TYPE(int32(s) & LOG_TYPE_MASK) {
	case SHORT_FILE_LOG_TYPE:
		return govStdShortLog
	case LONG_FILE_LOG_TYPE:
		return govStdLongLog
	}
	panic(fmt.Sprintf("no support std log:%d", s))
}

func (s std) Debug(v ...interface{}) {
	govLog(s).Output(int(s)&DEPTH_MASK, fmt.Sprint(v...))
}

func (s std) Debugf(format string, v ...interface{}) {
	govLog(s).Output(int(s)&DEPTH_MASK, fmt.Sprintf(format+"\n", v...))
}

func (s std) Error(v ...interface{}) {
	govLog(s).Output(int(s)&DEPTH_MASK, fmt.Sprint(v...))
}

func (s std) Errorf(format string, v ...interface{}) {
	govLog(s).Output(int(s)&DEPTH_MASK, fmt.Sprintf(format+"\n", v...))
}

func (s std) Info(v ...interface{}) {
	govLog(s).Output(int(s)&DEPTH_MASK, fmt.Sprint(v...))
}

func (s std) Infof(format string, v ...interface{}) {
	govLog(s).Output(int(s)&DEPTH_MASK, fmt.Sprintf(format+"\n", v...))
}
