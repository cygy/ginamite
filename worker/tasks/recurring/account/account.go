// Package account provides the recurring tasks about the accounts.
package account

import (
	"context"
	"time"

	"github.com/cygy/ginamite/common/queue"
	"github.com/cygy/ginamite/worker/tasks"

	"github.com/sirupsen/logrus"
)

// DeleteNeverUsed deletes the empty accounts which never logged in for intervalInDays days.
// The auth tokens revoked by the deletion are invalidated and the deletion is sent to the workers.
func DeleteNeverUsed(ctx context.Context, taskName string, intervalInDays uint, getNeverUsedAccounts func(context.Context, uint) []string, isUserEmpty func(context.Context, string) bool, deleteUser func(context.Context, string) map[string]time.Time) {
	userIDs := getNeverUsedAccounts(ctx, intervalInDays)
	countOfDeletedUsers := 0

	for _, userID := range userIDs {
		if isUserEmpty(ctx, userID) {
			deleteAndNotify(ctx, userID, deleteUser)
			countOfDeletedUsers++
		}
	}

	tasks.LogDone(taskName, &logrus.Fields{
		"count": countOfDeletedUsers,
	})
}

// DeleteInactive deletes the accounts inactive for intervalInMonths months.
// The auth tokens revoked by the deletion are invalidated and the deletion is sent to the workers.
func DeleteInactive(ctx context.Context, taskName string, intervalInMonths uint, getInactiveAccounts func(context.Context, uint) []string, deleteUser func(context.Context, string) map[string]time.Time) {
	userIDs := getInactiveAccounts(ctx, intervalInMonths)

	for _, userID := range userIDs {
		deleteAndNotify(ctx, userID, deleteUser)
	}

	tasks.LogDone(taskName, &logrus.Fields{
		"count": len(userIDs),
	})
}

// Sanitize deletes, or updates, the data associated to the deleted accounts.
func Sanitize(taskName string, sanitizeAccounts func() uint) {
	count := sanitizeAccounts()

	tasks.LogDone(taskName, &logrus.Fields{
		"count": count,
	})
}

// deleteAndNotify deletes a user, invalidates the auth tokens revoked and asks the workers to delete the user.
func deleteAndNotify(ctx context.Context, userID string, deleteUser func(context.Context, string) map[string]time.Time) {
	if revokedAuthTokens := deleteUser(ctx, userID); len(revokedAuthTokens) > 0 {
		queue.InvalidAuthTokens(revokedAuthTokens)
	}

	queue.DeleteUser(queue.TaskDeleteUser{
		UserID: userID,
	})
}
