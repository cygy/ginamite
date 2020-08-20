package request

import (
	"encoding/json"
	"strconv"
	"strings"

	"github.com/cygy/ginamite/api/context"
	"github.com/cygy/ginamite/api/response"
	"github.com/cygy/ginamite/common/config"
	"github.com/cygy/ginamite/common/localization"

	"github.com/gin-gonic/gin"
)

// DecodeBody : decodes the JSON-formatted body of a request to a struct.
func DecodeBody(c *gin.Context, v interface{}) {
	json.NewDecoder(c.Request.Body).Decode(&v)
}

// Unescape : unescapes a string.
func Unescape(s string) string {
	return strings.Replace(s, "\\\\", "\\", -1)
}

// GetOffsetAndLimitFromRequest : extracts the offset and limit from the GET parameters.
func GetOffsetAndLimitFromRequest(c *gin.Context, defaultOffset, defaultLimit, maxLimit int) (bool, int, int) {
	offset, err := strconv.Atoi(c.Query(OffsetRequestParameterName))
	if err != nil {
		offset = defaultOffset
	}

	limit, err := strconv.Atoi(c.Query(LimitRequestParameterName))
	if err != nil {
		limit = defaultLimit
	}

	locale := context.GetLocale(c)
	t := localization.Translate(locale)

	if offset < 0 {
		message := t("error.request.parameter.offset.negative.message")
		reason := t("error.request.parameter.offset.negative.reason")
		recovery := t("error.request.parameter.offset.negative.recovery")

		response.InvalidParameterValue(c, OffsetRequestParameterName, message, reason, recovery)
		return false, 0, 0
	}

	if limit < 0 {
		message := t("error.request.parameter.limit.negative.message")
		reason := t("error.request.parameter.limit.negative.reason")
		recovery := t("error.request.parameter.limit.negative.recovery")

		response.InvalidParameterValue(c, LimitRequestParameterName, message, reason, recovery)
		return false, 0, 0
	}

	if limit > maxLimit {
		message := t("error.request.parameter.limit.max.message")
		reason := t("error.request.parameter.limit.max.reason", localization.H{"LimitMax": maxLimit})
		recovery := t("error.request.parameter.limit.max.recovery", localization.H{"LimitMax": maxLimit})

		response.InvalidParameterValue(c, LimitRequestParameterName, message, reason, recovery)
		return false, 0, 0
	}

	return true, offset, limit
}

// GetLocaleValueFromRequest : extracts the locale from the URL.
func GetLocaleValueFromRequest(c *gin.Context, parameterName, errorMessageKey string) (string, bool) {
	locale := context.GetLocale(c)
	localeParameterValue := c.Param(parameterName)

	// Locale is mandatory.
	if len(localeParameterValue) == 0 {
		response.NotFoundParameterLocale(c, locale, parameterName, errorMessageKey)
		return "", false
	}

	// The locale must be supported by the application.
	if !config.Main.SupportsLocale(localeParameterValue) {
		response.UnsupportedRequestLocale(c, locale, localeParameterValue, config.Main.SupportedLocales, errorMessageKey)
		return "", false
	}

	return strings.ToLower(localeParameterValue), true
}
