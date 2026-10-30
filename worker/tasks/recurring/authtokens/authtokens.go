// Package authtokens provides the recurring tasks about the auth tokens.
package authtokens

import (
	"context"

	"github.com/cygy/ginamite/worker/tasks"

	"github.com/sirupsen/logrus"
)

// DeleteExpired deletes the expired auth tokens.
func DeleteExpired(ctx context.Context, taskName string, deleteExpired func(ctx context.Context) int64) {
	count := deleteExpired(ctx)

	tasks.LogDone(taskName, &logrus.Fields{
		"count": count,
	})
}
