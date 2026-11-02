// Package request reads and validates the parameters, the body and the files of an HTTP request.
package request

import (
	"encoding/json"
	"strconv"
	"strings"

	"github.com/cygy/ginamite/common/ctxkey"

	ctxLocale "github.com/cygy/ginamite/api/context/locale"
	"github.com/cygy/ginamite/api/response"
	"github.com/cygy/ginamite/common/localization"
	"github.com/cygy/ginamite/common/log"

	"github.com/gin-gonic/gin"
)

// DecodeBody decodes the JSON-formatted body of a request to a struct.
func DecodeBody(c *gin.Context, v any) {
	if err := json.NewDecoder(c.Request.Body).Decode(&v); err != nil {
		log.Warn("unable to decode the request body", "error", err)
	}
}

// Unescape unescapes a string.
func Unescape(s string) string {
	return strings.ReplaceAll(s, "\\\\", "\\")
}

// GetOffsetAndLimit extracts the offset and limit from the GET parameters.
func GetOffsetAndLimit(c *gin.Context, defaultOffset, defaultLimit, maxLimit int) (bool, int, int) {
	offset, err := strconv.Atoi(c.Query(OffsetParameterName))
	if err != nil {
		offset = defaultOffset
	}

	limit, err := strconv.Atoi(c.Query(LimitParameterName))
	if err != nil {
		limit = defaultLimit
	}

	locale := ctxLocale.Get(c)
	t := ctxkey.Localization.Get(c).Translate(locale)

	if offset < 0 {
		message := t("error.request.parameter.offset.negative.message")
		reason := t("error.request.parameter.offset.negative.reason")
		recovery := t("error.request.parameter.offset.negative.recovery")

		response.InvalidParameterValue(c, OffsetParameterName, message, reason, recovery)
		return false, 0, 0
	}

	if limit < 0 {
		message := t("error.request.parameter.limit.negative.message")
		reason := t("error.request.parameter.limit.negative.reason")
		recovery := t("error.request.parameter.limit.negative.recovery")

		response.InvalidParameterValue(c, LimitParameterName, message, reason, recovery)
		return false, 0, 0
	}

	if limit > maxLimit {
		message := t("error.request.parameter.limit.max.message")
		reason := t("error.request.parameter.limit.max.reason", localization.H{"LimitMax": maxLimit})
		recovery := t("error.request.parameter.limit.max.recovery", localization.H{"LimitMax": maxLimit})

		response.InvalidParameterValue(c, LimitParameterName, message, reason, recovery)
		return false, 0, 0
	}

	return true, offset, limit
}

// GetLocaleValue extracts the locale from the URL.
func GetLocaleValue(c *gin.Context, parameterName, errorMessageKey string) (string, bool) {
	locale := ctxLocale.Get(c)
	localeParameterValue := c.Param(parameterName)

	// Locale is mandatory.
	if len(localeParameterValue) == 0 {
		response.NotFoundParameterLocale(c, locale, parameterName, errorMessageKey)
		return "", false
	}

	// The locale must be supported by the application.
	if !ctxkey.Config.Get(c).SupportsLocale(localeParameterValue) {
		response.UnsupportedRequestLocale(c, locale, localeParameterValue, ctxkey.Config.Get(c).SupportedLocales, errorMessageKey)
		return "", false
	}

	return strings.ToLower(localeParameterValue), true
}

// GetOffset returns the value of the request parameter "offset".
func GetOffset(c *gin.Context) int64 {
	value, ok := c.Get(storedOffsetRequestParameterName)
	if !ok {
		return 0
	}

	return value.(int64)
}

// GetLimit returns the value of the request parameter "limit".
func GetLimit(c *gin.Context) int64 {
	value, ok := c.Get(storedLimitRequestParameterName)
	if !ok {
		return 0
	}

	return value.(int64)
}
