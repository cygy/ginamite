package context

import (
	"github.com/gin-gonic/gin"
	jwt "github.com/golang-jwt/jwt"
)

// Annex functions.
func getJWTToken(c *gin.Context) *jwt.Token {
	if value, ok := c.Get(tokenKey); ok {
		return value.(*jwt.Token)
	}

	return nil
}
