package tasks

import (
	"github.com/cygy/ginamite/common/log"
	"github.com/sirupsen/logrus"
)

// LogStarted : helper function to log the tasks.
func LogStarted(task string, fields *logrus.Fields) {
	entry := log.WithField("task", task)

	if fields != nil {
		entry = entry.WithFields(*fields)
	}

	entry.Info("task started")
}

// LogDone : helper function to log the tasks.
func LogDone(task string, fields *logrus.Fields) {
	entry := log.WithField("task", task)

	if fields != nil {
		entry = entry.WithFields(*fields)
	}

	entry.Info("task done")
}

// LogError : helper function to log the tasks.
func LogError(task string, err error, fields *logrus.Fields) {
	entry := log.WithField("task", task).WithField("error", err.Error())

	if fields != nil {
		entry = entry.WithFields(*fields)
	}

	entry.Error("task error")
}
