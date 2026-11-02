package terms

import (
	"context"

	"github.com/gin-gonic/gin"
)

// storageKey is the key of the storage in the Gin context.
const storageKey = "_terms_storage"

// Storage is the storage of the terms accepted by the users, provided by the application, such as the one of the
// storage plugin. The context is the one of the request.
type Storage interface {
	// UpdateVersionOfTermsAcceptedByUser saves the latest version of the terms accepted by the user.
	UpdateVersionOfTermsAcceptedByUser(ctx context.Context, userID, version string) error
	// GetLatestVersionOfTermsAcceptedByUser returns the latest version of the terms accepted by the user, or an empty
	// string if the user is unknown or accepted none.
	GetLatestVersionOfTermsAcceptedByUser(ctx context.Context, userID string) (string, error)
}

// Inject returns a middleware saving the storage in the Gin context, see StorageOf.
func Inject(storage Storage) gin.HandlerFunc {
	return func(c *gin.Context) {
		SetStorage(c, storage)
		c.Next()
	}
}

// SetStorage saves the storage in the Gin context, like Inject does. It is meant for the tests.
func SetStorage(c *gin.Context, storage Storage) {
	c.Set(storageKey, storage)
}

// StorageOf returns the storage saved in the Gin context by Inject, or nil if the application provides none: the
// feature relying on it is disabled.
func StorageOf(c *gin.Context) Storage {
	storage, _ := c.Get(storageKey)
	s, _ := storage.(Storage)

	return s
}
