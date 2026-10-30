// Package log provides the logger of the application and the request logging middleware.
package log

import (
	"os"

	"github.com/sirupsen/logrus"
)

// Main is the main logger of the application, set by Initialize.
var Main *logrus.Entry

// Initialize creates the main logger Main, writing to the standard output.
// The output is JSON in production and text otherwise.
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

// Debug logs a message at the debug level with the main logger.
func Debug(args ...any) {
	Main.Debug(args...)
}

// Error logs a message at the error level with the main logger.
func Error(args ...any) {
	Main.Error(args...)
}

// Fatal logs a message at the fatal level with the main logger, then exits the process.
func Fatal(args ...any) {
	Main.Fatal(args...)
}

// Info logs a message at the info level with the main logger.
func Info(args ...any) {
	Main.Info(args...)
}

// Panic logs a message at the panic level with the main logger, then panics.
func Panic(args ...any) {
	Main.Panic(args...)
}

// Print logs a message with the main logger.
func Print(args ...any) {
	Main.Print(args...)
}

// Warn logs a message at the warning level with the main logger.
func Warn(args ...any) {
	Main.Warn(args...)
}

// Warning logs a message at the warning level with the main logger.
func Warning(args ...any) {
	Main.Warning(args...)
}

// WithField returns an entry of the main logger with the field.
func WithField(key string, value any) *logrus.Entry {
	return Main.WithField(key, value)
}

// WithFields returns an entry of the main logger with the fields.
func WithFields(fields logrus.Fields) *logrus.Entry {
	return Main.WithFields(fields)
}
