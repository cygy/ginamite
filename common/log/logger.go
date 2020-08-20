package log

import (
	"os"

	"github.com/sirupsen/logrus"
)

// Initialize initialize the main logger of the application.
func Initialize(serverName string, productionEnvironment bool) {
	log := logrus.New()
	log.SetOutput(os.Stdout)

	if productionEnvironment {
		log.Formatter = &logrus.JSONFormatter{}
	} else {
		log.Formatter = &logrus.TextFormatter{}
	}

	Main = log.WithFields(logrus.Fields{
		"server": serverName,
	})
}

// Debug shorthand
func Debug(args ...interface{}) {
	Main.Debug(args...)
}

// Error shorthand
func Error(args ...interface{}) {
	Main.Error(args...)
}

// Fatal shorthand
func Fatal(args ...interface{}) {
	Main.Fatal(args...)
}

// Info shorthand
func Info(args ...interface{}) {
	Main.Info(args...)
}

// Panic shorthand
func Panic(args ...interface{}) {
	Main.Panic(args...)
}

// Print shorthand
func Print(args ...interface{}) {
	Main.Print(args...)
}

// Warn shorthand
func Warn(args ...interface{}) {
	Main.Warn(args...)
}

// Warning shorthand
func Warning(args ...interface{}) {
	Main.Warning(args...)
}

// WithField shorthand
func WithField(key string, value interface{}) *logrus.Entry {
	return Main.WithField(key, value)
}

// WithFields shorthand
func WithFields(fields logrus.Fields) *logrus.Entry {
	return Main.WithFields(fields)
}
