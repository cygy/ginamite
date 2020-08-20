package middleware

import (
	"github.com/cygy/ginamite/api/context"
	"github.com/cygy/ginamite/api/request"
	"github.com/cygy/ginamite/api/response"
	"github.com/cygy/ginamite/common/localization"
	"github.com/gin-gonic/gin"
	"github.com/globalsign/mgo/bson"
)

// VerifyURLParameter : verify that a URL parameter is present and its format.
func VerifyURLParameter(cfg Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		locale := context.GetLocale(c)
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

		if !cfg.SkipVerifyingInvalidObjectID && !bson.IsObjectIdHex(value) {
			response.InvalidRequestParameterWithDetails(c,
				t(cfg.MessageKey),
				t(cfg.InvalidObjectIDReasonKey),
				t(cfg.InvalidObjectIDRecoveryKey),
			)
			return
		}

		c.Set(cfg.StoreKey, value)
	}
}

// VerifyLocaleParameter : verify the parameter "locale".
func VerifyLocaleParameter(cfg Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		value, ok := request.GetLocaleValueFromRequest(c, cfg.Name, cfg.MessageKey)
		if !ok {
			return
		}

		c.Set(cfg.StoreKey, value)
	}
}
