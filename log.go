package main

import (
	"os"

	"runtime"
	"strings"

	"github.com/sirupsen/logrus"
)

var logger = &logrus.Entry{}

func initLog(debug bool) {
	logger = newLogger(debug)
}

func newLogger(debug bool) *logrus.Entry {
	l := logrus.Logger{
		Out:       os.Stdout,
		Formatter: new(logrus.TextFormatter),
		Hooks:     make(logrus.LevelHooks),
		Level:     logrus.InfoLevel,
	}

	if debug == true {
		l.SetLevel(logrus.DebugLevel)
	}

	entry := logrus.NewEntry(&l)
	return entry
}

// decorateRuntimeContext appends line, file and function context to the logger
func log() *logrus.Entry {
	if pc, fileFullPath, line, ok := runtime.Caller(1); ok {
		var fName string
		fNameFull := runtime.FuncForPC(pc).Name()
		if strings.Contains(fNameFull, ".") {
			splitfName := strings.Split(fNameFull, ".")
			fName = splitfName[len(splitfName)-1]
		} else {
			fName = fNameFull
		}

		split := strings.Split(fileFullPath, "/")
		file := split[len(split)-1]

		return logger.WithField("file", file).WithField("line", line).WithField("func", fName)
	} else {
		return logger
	}
}
