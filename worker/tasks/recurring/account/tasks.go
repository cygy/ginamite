package account

import (
	"time"

	"github.com/cygy/ginamite/common/queue"
	"github.com/cygy/ginamite/worker/tasks"
	"github.com/sirupsen/logrus"
)

// DeleteNeverUsed : deletes the never used accounts (never logged in).
func DeleteNeverUsed(taskName string, intervalInDays uint, getNeverUsedAccounts func(uint) []string, deleteUser func(userID string) map[string]time.Time) {
	userIDs := getNeverUsedAccounts(intervalInDays)

	// Delete the user.
	for _, userID := range userIDs {
		deleteUser(userID)

		queue.DeleteUser(queue.TaskDeleteUser{
			UserID: userID,
		})
	}

	tasks.LogDone(taskName, &logrus.Fields{
		"count": len(userIDs),
	})
}

// DeleteInactive : deletes the inactive accounts.
func DeleteInactive(taskName string, intervalInMonths uint, getInactiveAccounts func(uint) []string, deleteUser func(userID string) map[string]time.Time) {
	userIDs := getInactiveAccounts(intervalInMonths)

	// Delete the user.
	for _, userID := range userIDs {
		deleteUser(userID)

		queue.DeleteUser(queue.TaskDeleteUser{
			UserID: userID,
		})
	}

	tasks.LogDone(taskName, &logrus.Fields{
		"count": len(userIDs),
	})
}
