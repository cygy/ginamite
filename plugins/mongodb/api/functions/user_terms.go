package functions

import (
	"github.com/cygy/ginamite/common/mongo/session"
	"github.com/cygy/ginamite/plugins/mongodb/api/context"

	"github.com/gin-gonic/gin"
)

// UpdateVersionOfTermsAcceptedByUser : the user has accepted this version of terms.
func UpdateVersionOfTermsAcceptedByUser(c *gin.Context, userID, termsVersion string) error {
	mongoSession := session.Get(c)
	user := context.GetUser(c)
	return user.AcceptVersionOfTerms(termsVersion, mongoSession)
}
