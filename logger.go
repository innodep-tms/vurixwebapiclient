package vurixwebapiclient

import (
    "log"
    "os"
)

type Logger interface {
    Infof(format string, v ...interface{})
	Info(v ...interface{})
	Errorf(format string, v ...interface{})
	Error(v ...interface{})
	Warnf(format string, v ...interface{})
	Warn(v ...interface{})
	Debugf(format string, v ...interface{})
	Debug(v ...interface{})
}

func createLogger() *logger {
	l := &logger{l: log.New(os.Stderr, "", log.Ldate|log.Lmicroseconds)}
	return l
}

var _ Logger = (*logger)(nil)

type logger struct {
	l *log.Logger
}

func (l *logger) Infof(format string, v ...interface{}) {
	l.outputf("[INFO] "+format, v...)
}
func (l *logger) Info(v ...interface{}) {
	l.output("[INFO] ", v...)
}

func (l *logger) Errorf(format string, v ...interface{}) {
	l.outputf("[ERROR] "+format, v...)
}
func (l *logger) Error(v ...interface{}) {
	l.output("[ERROR] ", v...)
}

func (l *logger) Warnf(format string, v ...interface{}) {
	l.outputf("[WARN] "+format, v...)
}
func (l *logger) Warn(v ...interface{}) {
	l.output("[WARN] ", v...)
}

func (l *logger) Debugf(format string, v ...interface{}) {
	l.outputf("[DEBUG] "+format, v...)
}
func (l *logger) Debug(v ...interface{}) {
	l.output("[DEBUG] ", v...)
}

func (l *logger) outputf(format string, v ...interface{}) {
	if len(v) == 0 {
		l.l.Print(format)
		return
	}
	l.l.Printf(format, v...)
}
func (l *logger) output(format string, v ...interface{}) {
	l.l.Print(append([]interface{}{format}, v...))
}