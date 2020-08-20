package authtokens

import (
	"github.com/cygy/ginamite/worker/tasks"
	"github.com/sirupsen/logrus"
)

// DeleteExpired : clears the expired auth tokens.
func DeleteExpired(taskName string, deleteExpired func() uint) {
	count := deleteExpired()

	tasks.LogDone(taskName, &logrus.Fields{
		"count": count,
	})
}
