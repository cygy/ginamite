package contact

import (
	"context"

	"github.com/gin-gonic/gin"
)

// storageKey is the key of the storage in the Gin context.
const storageKey = "_contact_storage"

// Storage is the storage of the contacts, provided by the application, such as the one of the storage plugin. The
// context is the one of the request.
type Storage interface {
	// CreateContact saves a message sent by a visitor.
	CreateContact(ctx context.Context, emailAddress, subject, text string) error
	// UpdateContact marks a contact as done, or not.
	UpdateContact(ctx context.Context, contactID string, done bool) error
	// GetContacts returns a page of contacts, for the administrators.
	GetContacts(ctx context.Context, offset, limit int64) (any, error)
	// GetContact returns a contact, for the administrators.
	GetContact(ctx context.Context, contactID string) (any, error)
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
