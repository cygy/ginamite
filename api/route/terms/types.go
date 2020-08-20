package terms

import "github.com/gin-gonic/gin"

// UpdateVersionOfTermsAcceptedByUserFunc : function to save the latest version of the terms acepted by the user.
type UpdateVersionOfTermsAcceptedByUserFunc func(c *gin.Context, userID, version string) error
