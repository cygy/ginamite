// Package tasks provides the helpers shared by the tasks of the worker.
package tasks

import (
	"github.com/cygy/ginamite/common/log"

	"github.com/sirupsen/logrus"
)

// LogStarted logs that a task started, with the optional fields.
func LogStarted(task string, fields *logrus.Fields) {
	entry := log.WithField("task", task)

	if fields != nil {
		entry = entry.WithFields(*fields)
	}

	entry.Info("task started")
}

// LogDone logs that a task is done, with the optional fields.
func LogDone(task string, fields *logrus.Fields) {
	entry := log.WithField("task", task)

	if fields != nil {
		entry = entry.WithFields(*fields)
	}

	entry.Info("task done")
}

// LogError logs the error of a task, with the optional fields.
func LogError(task string, err error, fields *logrus.Fields) {
	entry := log.WithField("task", task)
	if err != nil {
		entry = entry.WithField("error", err.Error())
	}

	if fields != nil {
		entry = entry.WithFields(*fields)
	}

	entry.Error("task error")
}
