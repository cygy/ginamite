package middleware

import (
	"github.com/cygy/ginamite/api/request"
	"github.com/gin-gonic/gin"
)

// VerifyLimitAndOffset : verify the request parameters "limit" and "offset".
func VerifyLimitAndOffset(defaultOffset, defaultLimit, maxLimit int) gin.HandlerFunc {
	return func(c *gin.Context) {
		ok, offset, limit := request.GetOffsetAndLimitFromRequest(c, defaultOffset, defaultLimit, maxLimit)
		if !ok {
			return
		}

		c.Set("_offset", offset)
		c.Set("_limit", limit)
	}
}

// GetOffsetFromRequest : returns the value of the request parameter "offset".
func GetOffsetFromRequest(c *gin.Context) int {
	value, ok := c.Get("_offset")
	if !ok {
		return 0
	}

	return value.(int)
}

// GetLimitFromRequest : returns the value of the request parameter "limit".
func GetLimitFromRequest(c *gin.Context) int {
	value, ok := c.Get("_limit")
	if !ok {
		return 0
	}

	return value.(int)
}
