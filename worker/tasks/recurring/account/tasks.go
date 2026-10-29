package account

import (
	"context"
	"time"

	"github.com/cygy/ginamite/common/queue"
	"github.com/cygy/ginamite/worker/tasks"
	"github.com/sirupsen/logrus"
)

// DeleteNeverUsed : deletes the never used accounts (never logged in).
func DeleteNeverUsed(ctx context.Context, taskName string, intervalInDays uint, getNeverUsedAccounts func(context.Context, uint) []string, isUserEmpty func(context.Context, string) bool, deleteUser func(context.Context, string) map[string]time.Time) {
	userIDs := getNeverUsedAccounts(ctx, intervalInDays)
	countOfDeletedUsers := 0

	// Delete the user.
	for _, userID := range userIDs {
		if isUserEmpty(ctx, userID) {
			deleteUser(ctx, userID)
			countOfDeletedUsers++

			queue.DeleteUser(queue.TaskDeleteUser{
				UserID: userID,
			})
		}
	}

	tasks.LogDone(taskName, &logrus.Fields{
		"count": countOfDeletedUsers,
	})
}

// DeleteInactive : deletes the inactive accounts.
func DeleteInactive(ctx context.Context, taskName string, intervalInMonths uint, getInactiveAccounts func(context.Context, uint) []string, deleteUser func(context.Context, string) map[string]time.Time) {
	userIDs := getInactiveAccounts(ctx, intervalInMonths)

	// Delete the user.
	for _, userID := range userIDs {
		deleteUser(ctx, userID)

		queue.DeleteUser(queue.TaskDeleteUser{
			UserID: userID,
		})
	}

	tasks.LogDone(taskName, &logrus.Fields{
		"count": len(userIDs),
	})
}

// Sanitize : deletes/updates the data associated to deleted accounts.
func Sanitize(taskName string, sanitizeAccounts func() uint) {
	count := sanitizeAccounts()

	tasks.LogDone(taskName, &logrus.Fields{
		"count": count,
	})
}
