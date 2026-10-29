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
func Debug(args ...any) {
	Main.Debug(args...)
}

// Error shorthand
func Error(args ...any) {
	Main.Error(args...)
}

// Fatal shorthand
func Fatal(args ...any) {
	Main.Fatal(args...)
}

// Info shorthand
func Info(args ...any) {
	Main.Info(args...)
}

// Panic shorthand
func Panic(args ...any) {
	Main.Panic(args...)
}

// Print shorthand
func Print(args ...any) {
	Main.Print(args...)
}

// Warn shorthand
func Warn(args ...any) {
	Main.Warn(args...)
}

// Warning shorthand
func Warning(args ...any) {
	Main.Warning(args...)
}

// WithField shorthand
func WithField(key string, value any) *logrus.Entry {
	return Main.WithField(key, value)
}

// WithFields shorthand
func WithFields(fields logrus.Fields) *logrus.Entry {
	return Main.WithFields(fields)
}
