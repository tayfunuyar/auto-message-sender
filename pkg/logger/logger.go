package logger

import (
	"os"
	"sync"

	"github.com/sirupsen/logrus"
)

const (
	DebugLevel = logrus.DebugLevel
	InfoLevel  = logrus.InfoLevel
	WarnLevel  = logrus.WarnLevel
	ErrorLevel = logrus.ErrorLevel
	FatalLevel = logrus.FatalLevel
)

var (
	log  *logrus.Logger
	once sync.Once
)

func Init(level logrus.Level) {
	once.Do(func() {
		log = logrus.New()
		log.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: "2006-01-02T15:04:05.000Z07:00",
		})
		log.SetOutput(os.Stdout)
		log.SetLevel(level)
	})
}

func GetLogger() *logrus.Logger {
	if log == nil {
		Init(InfoLevel)
	}
	return log
}

func Debug(msg string) {
	GetLogger().Debug(msg)
}

func Debugf(format string, args ...interface{}) {
	GetLogger().Debugf(format, args...)
}

func Info(msg string) {
	GetLogger().Info(msg)
}

func Infof(format string, args ...interface{}) {
	GetLogger().Infof(format, args...)
}

func Warn(msg string) {
	GetLogger().Warn(msg)
}

func Warnf(format string, args ...interface{}) {
	GetLogger().Warnf(format, args...)
}

func Error(msg string) {
	GetLogger().Error(msg)
}

func Errorf(format string, args ...interface{}) {
	GetLogger().Errorf(format, args...)
}

func Fatal(msg string) {
	GetLogger().Fatal(msg)
}

func Fatalf(format string, args ...interface{}) {
	GetLogger().Fatalf(format, args...)
}

func WithField(key string, value interface{}) *logrus.Entry {
	return GetLogger().WithField(key, value)
}

func WithFields(fields logrus.Fields) *logrus.Entry {
	return GetLogger().WithFields(fields)
}

func WithError(err error) *logrus.Entry {
	return GetLogger().WithError(err)
}
