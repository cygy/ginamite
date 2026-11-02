// Package tasks provides the helpers shared by the tasks of the worker.
package tasks

import (
	"github.com/cygy/ginamite/common/log"
)

// LogStarted logs that a task started. The arguments are alternating keys and values, added to the record.
func LogStarted(task string, args ...any) {
	log.With(append([]any{"task", task}, args...)...).Info("task started")
}

// LogDone logs that a task is done. The arguments are alternating keys and values, added to the record.
func LogDone(task string, args ...any) {
	log.With(append([]any{"task", task}, args...)...).Info("task done")
}

// LogError logs the error of a task, if there is one. The arguments are alternating keys and values, added to the
// record.
func LogError(task string, err error, args ...any) {
	attrs := []any{"task", task}
	if err != nil {
		attrs = append(attrs, "error", err.Error())
	}

	log.With(append(attrs, args...)...).Error("task error")
}
