package functions

import (
	"context"

	"github.com/cygy/ginamite/common/log"

	"github.com/cygy/ginamite/common/mongo/session"
	"github.com/cygy/ginamite/plugins/mongodb/model"
)

// DeleteExpiredTokens deletes the expired auth tokens and returns the number of tokens deleted.
func DeleteExpiredTokens(ctx context.Context) int64 {
	db := session.Database()
	count, err := model.DeleteExpiredAuthTokens(ctx, db)
	if err != nil {
		log.WithField("error", err).Error("unable to delete the expired auth tokens")
	}

	return count
}
