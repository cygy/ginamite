package context

import (
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

// Annex functions.
func getJWTToken(c *gin.Context) *jwt.Token {
	if value, ok := c.Get(tokenKey); ok {
		return value.(*jwt.Token)
	}

	return nil
}
