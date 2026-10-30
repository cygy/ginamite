package request

import (
	ctxLocale "github.com/cygy/ginamite/api/context/locale"
	"github.com/cygy/ginamite/api/response"
	"github.com/cygy/ginamite/common/localization"
	"github.com/cygy/ginamite/common/mongo/bson"
	"github.com/gin-gonic/gin"
)

// VerifyLimitAndOffset verify the request parameters "limit" and "offset".
func VerifyLimitAndOffset(defaultOffset, defaultLimit, maxLimit int) gin.HandlerFunc {
	return func(c *gin.Context) {
		ok, offset, limit := GetOffsetAndLimit(c, defaultOffset, defaultLimit, maxLimit)
		if !ok {
			return
		}

		c.Set(storedOffsetRequestParameterName, offset)
		c.Set(storedLimitRequestParameterName, limit)
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

// VerifyURLParameter verify that a URL parameter is present and its format.
func VerifyURLParameter(cfg Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		locale := ctxLocale.Get(c)
		t := localization.Translate(locale)

		value := c.Param(cfg.Name)
		if len(value) == 0 {
			response.InvalidRequestParameterWithDetails(c,
				t(cfg.MessageKey),
				t(cfg.NotFoundReasonKey),
				t(cfg.NotFoundRecoveryKey),
			)
			return
		}

		if !cfg.SkipVerifyingInvalidObjectID {
			if _, err := bson.ObjectIDFromHex(value); err != nil {
				response.InvalidRequestParameterWithDetails(c,
					t(cfg.MessageKey),
					t(cfg.InvalidObjectIDReasonKey),
					t(cfg.InvalidObjectIDRecoveryKey),
				)
				return
			}
		}

		c.Set(cfg.StoreKey, value)
	}
}
