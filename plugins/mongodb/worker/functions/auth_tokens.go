package functions

import (
	"context"

	"github.com/cygy/ginamite/common/mongo/session"
	"github.com/cygy/ginamite/plugins/mongodb/model"
)

// DeleteExpiredTokens : delete the expired tokens.
func DeleteExpiredTokens(ctx context.Context) int64 {
	db := session.Database()
	count, _ := model.DeleteExpiredAuthTokens(ctx, db)
	return count
}
