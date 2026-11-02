// Package authtokens provides the recurring tasks about the auth tokens.
package authtokens

import (
	"context"
	"fmt"

	"github.com/cygy/ginamite/worker/tasks"

	"github.com/sirupsen/logrus"
)

// DeleteExpired deletes the expired auth tokens.
// The error is logged if it fails.
func DeleteExpired(ctx context.Context, taskName string, deleteExpired func(ctx context.Context) (int64, error)) {
	count, err := deleteExpired(ctx)
	if err != nil {
		tasks.LogError(taskName, fmt.Errorf("unable to delete the expired auth tokens: %w", err), &logrus.Fields{"count": count})

		return
	}

	tasks.LogDone(taskName, &logrus.Fields{
		"count": count,
	})
}
