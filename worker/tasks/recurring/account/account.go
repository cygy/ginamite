// Package account provides the recurring tasks about the accounts.
package account

import (
	"context"
	"fmt"
	"time"

	"github.com/cygy/ginamite/common/log"
	"github.com/cygy/ginamite/common/queue"
	"github.com/cygy/ginamite/worker/tasks"

	"github.com/sirupsen/logrus"
)

// DeleteNeverUsed deletes the empty accounts which never logged in for intervalInDays days.
// The auth tokens revoked by the deletion are invalidated and the deletion is sent to the workers.
func DeleteNeverUsed(ctx context.Context, q *queue.Queue, taskName string, intervalInDays uint, getNeverUsedAccounts func(context.Context, uint) ([]string, error), isUserEmpty func(context.Context, string) (bool, error), deleteUser func(context.Context, string) (map[string]time.Time, error)) {
	userIDs, err := getNeverUsedAccounts(ctx, intervalInDays)
	if err != nil {
		tasks.LogError(taskName, fmt.Errorf("unable to get the accounts never used: %w", err), nil)

		return
	}

	countOfDeletedUsers := 0
	for _, userID := range userIDs {
		empty, err := isUserEmpty(ctx, userID)
		if err != nil {
			log.WithFields(logrus.Fields{"user": userID, "error": err.Error()}).Error("unable to know if the user is empty")

			continue
		}

		if empty && deleteAndNotify(ctx, q, userID, deleteUser) {
			countOfDeletedUsers++
		}
	}

	tasks.LogDone(taskName, &logrus.Fields{
		"count": countOfDeletedUsers,
	})
}

// DeleteInactive deletes the accounts inactive for intervalInMonths months.
// The auth tokens revoked by the deletion are invalidated and the deletion is sent to the workers.
func DeleteInactive(ctx context.Context, q *queue.Queue, taskName string, intervalInMonths uint, getInactiveAccounts func(context.Context, uint) ([]string, error), deleteUser func(context.Context, string) (map[string]time.Time, error)) {
	userIDs, err := getInactiveAccounts(ctx, intervalInMonths)
	if err != nil {
		tasks.LogError(taskName, fmt.Errorf("unable to get the inactive accounts: %w", err), nil)

		return
	}

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
// The error is logged if it fails.
func Sanitize(ctx context.Context, taskName string, sanitizeAccounts func(context.Context) (uint, error)) {
	count, err := sanitizeAccounts(ctx)
	if err != nil {
		tasks.LogError(taskName, fmt.Errorf("unable to sanitize the accounts: %w", err), &logrus.Fields{"count": count})

		return
	}

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
