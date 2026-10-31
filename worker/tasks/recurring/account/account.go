// Package account provides the recurring tasks about the accounts.
package account

import (
	"context"
	"time"

	"github.com/cygy/ginamite/common/log"
	"github.com/cygy/ginamite/common/queue"
	"github.com/cygy/ginamite/worker/tasks"

	"github.com/sirupsen/logrus"
)

// DeleteNeverUsed deletes the empty accounts which never logged in for intervalInDays days.
// The auth tokens revoked by the deletion are invalidated and the deletion is sent to the workers.
func DeleteNeverUsed(ctx context.Context, q *queue.Queue, taskName string, intervalInDays uint, getNeverUsedAccounts func(context.Context, uint) []string, isUserEmpty func(context.Context, string) bool, deleteUser func(context.Context, string) (map[string]time.Time, error)) {
	userIDs := getNeverUsedAccounts(ctx, intervalInDays)
	countOfDeletedUsers := 0

	for _, userID := range userIDs {
		if isUserEmpty(ctx, userID) && deleteAndNotify(ctx, q, userID, deleteUser) {
			countOfDeletedUsers++
		}
	}

	tasks.LogDone(taskName, &logrus.Fields{
		"count": countOfDeletedUsers,
	})
}

// DeleteInactive deletes the accounts inactive for intervalInMonths months.
// The auth tokens revoked by the deletion are invalidated and the deletion is sent to the workers.
func DeleteInactive(ctx context.Context, q *queue.Queue, taskName string, intervalInMonths uint, getInactiveAccounts func(context.Context, uint) []string, deleteUser func(context.Context, string) (map[string]time.Time, error)) {
	userIDs := getInactiveAccounts(ctx, intervalInMonths)

	countOfDeletedUsers := 0
	for _, userID := range userIDs {
		if deleteAndNotify(ctx, q, userID, deleteUser) {
			countOfDeletedUsers++
		}
	}

	tasks.LogDone(taskName, &logrus.Fields{
		"count": countOfDeletedUsers,
	})
}

// Sanitize deletes, or updates, the data associated to the deleted accounts.
func Sanitize(taskName string, sanitizeAccounts func() uint) {
	count := sanitizeAccounts()

	tasks.LogDone(taskName, &logrus.Fields{
		"count": count,
	})
}

// deleteAndNotify deletes a user, invalidates the auth tokens revoked and asks the workers to delete the user. It
// logs the error, and reports false, if the user can not be deleted: the next run tries again.
func deleteAndNotify(ctx context.Context, q *queue.Queue, userID string, deleteUser func(context.Context, string) (map[string]time.Time, error)) bool {
	revokedAuthTokens, err := deleteUser(ctx, userID)
	if err != nil {
		log.WithFields(logrus.Fields{"user": userID, "error": err.Error()}).Error("unable to delete the user")

		return false
	}

	if len(revokedAuthTokens) > 0 {
		q.InvalidAuthTokens(revokedAuthTokens)
	}

	q.DeleteUser(queue.TaskDeleteUser{
		UserID: userID,
	})

	return true
}
