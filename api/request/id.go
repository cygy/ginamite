package request

import "github.com/gin-gonic/gin"

// idValidatorKey is the key of the validator of the identifiers in the gin context.
const idValidatorKey = "_idvalidator"

// IDValidatorFunc reports whether an identifier is well-formed.
type IDValidatorFunc func(id string) bool

// InjectIDValidator returns a middleware saving in the gin context the function validating the identifiers, see
// IsValidID. The format of the identifiers depends on the database, so the API server gives the function of the
// database plugin, Handlers.IsValidID.
func InjectIDValidator(validator IDValidatorFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		SetIDValidator(c, validator)
		c.Next()
	}
}

// SetIDValidator saves the function validating the identifiers in the gin context, like InjectIDValidator does. It is
// meant for the tests.
func SetIDValidator(c *gin.Context, validator IDValidatorFunc) {
	c.Set(idValidatorKey, validator)
}

// IsValidID reports whether id is a well-formed identifier of a stored entity, such as a notification, according to
// the validator of the gin context. Every identifier is valid if the context has no validator.
func IsValidID(c *gin.Context, id string) bool {
	if value, ok := c.Get(idValidatorKey); ok {
		return value.(IDValidatorFunc)(id)
	}

	return true
}
