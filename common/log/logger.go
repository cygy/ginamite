// Package log provides the logger of the application and the request logging middleware.
package log

import (
	"log/slog"
	"os"
)

// Main is the main logger of the application, set by Initialize.
var Main *slog.Logger

// Initialize creates the main logger Main, writing to the standard output. The output is JSON in production and text
// otherwise. Every record carries the name of the server.
func Initialize(serverName string, productionEnvironment bool) {
	var handler slog.Handler
	if productionEnvironment {
		handler = slog.NewJSONHandler(os.Stdout, nil)
	} else {
		handler = slog.NewTextHandler(os.Stdout, nil)
	}

	Main = slog.New(handler).With("server", serverName)
}

// Debug logs a message at the debug level with the main logger. The arguments are alternating keys and values, see
// slog.Logger.Log.
func Debug(msg string, args ...any) {
	Main.Debug(msg, args...)
}

// Info logs a message at the info level with the main logger.
func Info(msg string, args ...any) {
	Main.Info(msg, args...)
}

// Warn logs a message at the warning level with the main logger.
func Warn(msg string, args ...any) {
	Main.Warn(msg, args...)
}

// Error logs a message at the error level with the main logger.
func Error(msg string, args ...any) {
	Main.Error(msg, args...)
}

// With returns the main logger with the attributes added to each of its records.
func With(args ...any) *slog.Logger {
	return Main.With(args...)
}
