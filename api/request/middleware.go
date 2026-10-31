package request

import (
	ctxLocale "github.com/cygy/ginamite/api/context/locale"
	"github.com/cygy/ginamite/api/response"
	"github.com/cygy/ginamite/common/ctxlocalization"
	"github.com/gin-gonic/gin"
)

// VerifyLimitAndOffset verify the request parameters "limit" and "offset".
func VerifyLimitAndOffset(defaultOffset, defaultLimit, maxLimit int) gin.HandlerFunc {
	return func(c *gin.Context) {
		ok, offset, limit := GetOffsetAndLimit(c, defaultOffset, defaultLimit, maxLimit)
		if !ok {
			return
		}

		c.Set(storedOffsetRequestParameterName, int64(offset))
		c.Set(storedLimitRequestParameterName, int64(limit))
	}
}

// VerifyLocaleParameter verify the parameter "locale".
func VerifyLocaleParameter(cfg Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		value, ok := GetLocaleValue(c, cfg.Name, cfg.MessageKey)
		if !ok {
			return
		}

		c.Set(cfg.StoreKey, value)
	}
}

// VerifyURLParameter verify that a URL parameter is present and, unless the configuration skips it, that it is a valid
// identifier, see IsValidID.
func VerifyURLParameter(cfg Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		locale := ctxLocale.Get(c)
		t := ctxlocalization.Get(c).Translate(locale)

		value := c.Param(cfg.Name)
		if len(value) == 0 {
			response.InvalidRequestParameterWithDetails(c,
				t(cfg.MessageKey),
				t(cfg.NotFoundReasonKey),
				t(cfg.NotFoundRecoveryKey),
			)
			return
		}

		if !cfg.SkipVerifyingInvalidID {
			if !IsValidID(value) {
				response.InvalidRequestParameterWithDetails(c,
					t(cfg.MessageKey),
					t(cfg.InvalidIDReasonKey),
					t(cfg.InvalidIDRecoveryKey),
				)
				return
			}
		}

		c.Set(cfg.StoreKey, value)
	}
}
