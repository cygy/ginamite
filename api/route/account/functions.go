package account

import (
	"github.com/cygy/ginamite/api/validator"

	"github.com/gin-gonic/gin"
)

// storageKey is the key of the functions in the Gin context.
const storageKey = "_account_functions"

// Storage are the functions provided by the application, such as the ones of the storage plugin, used by the
// handlers of the package. A missing function disables the feature relying on it.
type Storage struct {
	// NewTokenValidator returns the validator of the tokens issued by a third-party authentication method. The default
	// one validates the tokens of Facebook and Google with their API.
	NewTokenValidator func(c *gin.Context, method string) validator.TokenValidator

	// Notifications is the storage of the notifications of the users.
	Notifications NotificationsStorage

	// Tokens is the storage of the auth tokens.
	Tokens TokensStorage

	// Processes is the storage of the processes bound to a user.
	Processes ProcessesStorage

	// Accounts is the storage of the accounts of the users.
	Accounts AccountsStorage
}

// Inject returns a middleware saving the functions in the Gin context, see StorageOf.
func Inject(functions *Storage) gin.HandlerFunc {
	return func(c *gin.Context) {
		SetStorage(c, functions)
		c.Next()
	}
}

// SetStorage saves the functions in the Gin context, like Inject does. It is meant for the tests.
func SetStorage(c *gin.Context, functions *Storage) {
	c.Set(storageKey, functions)
}

// StorageOf returns the functions saved in the Gin context by Inject.
// It panics if there are none: the handlers can not work without them.
func StorageOf(c *gin.Context) *Storage {
	value, ok := c.Get(storageKey)
	if !ok {
		panic("account: no functions in the context, add the middleware Inject")
	}

	return value.(*Storage)
}
