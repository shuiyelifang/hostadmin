package log

import (
	"github.com/sirupsen/logrus"
	// "io/ioutil"
	"os"
	"sync"
)

var logger *logrus.Logger
var initOnce sync.Once

func init() {
	logger = new(logrus.Logger)
	InitLogger("", false)
}

func GetLogger() *logrus.Logger {
	return logger
}

func InitLogger(filename string, debug bool) {
	// logFile := ioutil.Discard
	logFile := os.Stdout

	if filename != "" {
		var err error
		logFile, err = os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			panic(err)
		}
	}

	logger.Out = logFile
	logger.Formatter = new(logrus.TextFormatter)
	// logger.Formatter = new(logrus.JSONFormatter)

	if debug {
		logger.Level = logrus.DebugLevel
	} else {
		logger.Level = logrus.InfoLevel
	}
}
