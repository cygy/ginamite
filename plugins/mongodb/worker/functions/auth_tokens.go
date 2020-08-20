package functions

import (
	"github.com/cygy/ginamite/common/mongo/session"
	"github.com/cygy/ginamite/plugins/mongodb/model"
)

// DeleteExpiredTokens : delete the expired tokens.
func DeleteExpiredTokens() uint {
	session, db := session.Copy()
	defer session.Close()

	count, _ := model.DeleteExpiredAuthTokens(db)

	return uint(count)
}
