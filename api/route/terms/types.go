package terms

import "github.com/gin-gonic/gin"

// UpdateVersionOfTermsAcceptedByUserFunc : function to save the latest version of the terms acepted by the user.
type UpdateVersionOfTermsAcceptedByUserFunc func(c *gin.Context, userID, version string) error

// GetLatestVersionOfTermsAcceptedByUserFunc : returns the latest version of the terms accepted by the user.
type GetLatestVersionOfTermsAcceptedByUserFunc func(c *gin.Context, userID string) string
