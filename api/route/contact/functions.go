package contact

import "github.com/gin-gonic/gin"

// functionsKey is the key of the functions in the Gin context.
const functionsKey = "_contact_functions"

// Functions are the functions provided by the application, such as the ones of the storage plugin, used by the
// handlers of the package. A missing function disables the feature relying on it.
type Functions struct {
	// CreateContact is the function to create a contact.
	CreateContact CreateContactFunc

	// UpdateContact is the function to update a contact.
	UpdateContact UpdateContactFunc

	// GetContacts is the function to get the contacts.
	GetContacts GetContactsFunc

	// GetContact is the function to get a contact.
	GetContact GetContactFunc
}

// Inject returns a middleware saving the functions in the Gin context, see FunctionsOf.
func Inject(functions *Functions) gin.HandlerFunc {
	return func(c *gin.Context) {
		SetFunctions(c, functions)
		c.Next()
	}
}

// SetFunctions saves the functions in the Gin context, like Inject does. It is meant for the tests.
func SetFunctions(c *gin.Context, functions *Functions) {
	c.Set(functionsKey, functions)
}

// FunctionsOf returns the functions saved in the Gin context by Inject.
// It panics if there are none: the handlers can not work without them.
func FunctionsOf(c *gin.Context) *Functions {
	value, ok := c.Get(functionsKey)
	if !ok {
		panic("contact: no functions in the context, add the middleware Inject")
	}

	return value.(*Functions)
}
