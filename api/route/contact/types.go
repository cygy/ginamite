package contact

import "github.com/gin-gonic/gin"

// CreateContactFunc : function to create a contact.
type CreateContactFunc func(c *gin.Context, emailAddress, subject, text string) error

// UpdateContactFunc : function to update a contact.
type UpdateContactFunc func(c *gin.Context, contactID string, done bool) error

// GetContactsFunc : function to get the contacts.
type GetContactsFunc func(c *gin.Context, offset, limit int) (interface{}, error)

// GetContactFunc : function to get a contact.
type GetContactFunc func(c *gin.Context, contactID string) (interface{}, error)
