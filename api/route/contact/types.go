package contact

import "github.com/gin-gonic/gin"

// CreateContactFunc is a function to create a contact.
type CreateContactFunc func(c *gin.Context, emailAddress, subject, text string) error

// UpdateContactFunc is a function to update a contact.
type UpdateContactFunc func(c *gin.Context, contactID string, done bool) error

// GetContactsFunc is a function to get the contacts.
type GetContactsFunc func(c *gin.Context, offset, limit int64) (any, error)

// GetContactFunc is a function to get a contact.
type GetContactFunc func(c *gin.Context, contactID string) (any, error)
