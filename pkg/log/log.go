package log

import (
	"fmt"
	"io"
	"os"

	"github.com/sirupsen/logrus"
)

func InitializeLog(logLevel, logFile string) {
	logrus.SetFormatter(&logrus.TextFormatter{
		ForceColors:     true,
		TimestampFormat: "2008-01-01 12:00:00",
		FullTimestamp:   true,
	})

	switch logLevel {
	case "panic":
		logrus.SetLevel(logrus.PanicLevel)
	case "fatal":
		logrus.SetLevel(logrus.FatalLevel)
	case "error":
		logrus.SetLevel(logrus.ErrorLevel)
	case "warn":
		logrus.SetLevel(logrus.WarnLevel)
	case "info":
		logrus.SetLevel(logrus.InfoLevel)
	case "debug":
		logrus.SetLevel(logrus.DebugLevel)
	case "trace":
		logrus.SetLevel(logrus.TraceLevel)
	}

	if len(logFile) > 0 {
		file, err := os.OpenFile(logFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
		if err != nil {
			fmt.Printf("[ERROR] failed to open log file: %v\n", err)
		} else {
			logrus.SetOutput(io.MultiWriter(os.Stdout, file))
			logrus.RegisterExitHandler(func() {
				_ = file.Close()
			})
		}
	}
}

func Infof(format string, args ...interface{}) {
	logrus.Infof(format, args...)
}

func Warnf(format string, args ...interface{}) {
	logrus.Warnf(format, args...)
}

func Errorf(format string, args ...interface{}) {
	logrus.Errorf(format, args...)
}
