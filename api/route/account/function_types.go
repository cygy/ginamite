package account

import (
	"github.com/cygy/ginamite/api/validator"
	"github.com/cygy/ginamite/common/authentication"

	"github.com/gin-gonic/gin"
)

// GetUserDetailsBySocialNetworkIDFunc is a function to get the user details from its unique social network identifier.
type GetUserDetailsBySocialNetworkIDFunc func(c *gin.Context, userID string) (authentication.User, error)

// DoesUserWithSocialNetworkUserIDExistFunc is a function to know if a user with a social network user id is registered already.
type DoesUserWithSocialNetworkUserIDExistFunc func(c *gin.Context, userID string) (bool, error)

// SaveSocialNetworkUserTokenDetailsToUserFunc is a function to save the auth token details from a social network to an account.
type SaveSocialNetworkUserTokenDetailsToUserFunc func(c *gin.Context, userID string, tokenInfos validator.TokenInfos) error

// CreateNewProcessFunc is a function to start a new process bound to a user.
type CreateNewProcessFunc func(c *gin.Context, userID string) (processID, processKey string, err error)

// CreateNewProcessWithValueFunc is a function to start a new process bound to a user.
type CreateNewProcessWithValueFunc func(c *gin.Context, userID, value string) (processID, processKey string, err error)
