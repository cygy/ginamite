package response

import (
	"strings"

	"github.com/cygy/ginamite/common/errors"
	"github.com/cygy/ginamite/common/errors/auth"
	"github.com/cygy/ginamite/common/errors/database"
	"github.com/cygy/ginamite/common/errors/file"
	"github.com/cygy/ginamite/common/errors/locale"
	"github.com/cygy/ginamite/common/errors/request"
	"github.com/cygy/ginamite/common/localization"
	"github.com/cygy/ginamite/common/response"

	humanize "github.com/dustin/go-humanize"
	"github.com/gin-gonic/gin"
)

// H : shorthand to the hashmap type.
type H gin.H

// UnsupportedRequestLocale : returns a HTTP 400 error.
func UnsupportedRequestLocale(c *gin.Context, userLocale, usedLocale string, supportedLocales []string, messageKey string) {
	t := localization.Translate(userLocale)
	BadRequestWithError(c, response.Error{
		Domain:   locale.Domain,
		Code:     locale.UnsupportedLocale,
		Message:  t(messageKey),
		Reason:   t("error.request.parameter.locale.unsupported.reason", localization.H{"Locale": usedLocale}),
		Recovery: t("error.request.parameter.locale.unsupported.recovery", localization.H{"Locales": strings.Join(supportedLocales, ",")}),
	})
}

// UnsupportedParameterLocale : returns a HTTP 400 error.
func UnsupportedParameterLocale(c *gin.Context, userLocale, usedLocale, parameterName string, supportedLocales []string, messageKey string) {
	t := localization.Translate(userLocale)
	InvalidParameterValue(c, parameterName,
		t(messageKey),
		t("error.request.parameter.locale.unsupported.reason", localization.H{"Locale": usedLocale}),
		t("error.request.parameter.locale.unsupported.recovery", localization.H{"Locales": strings.Join(supportedLocales, ",")}),
	)
}

// NotFoundParameterLocale : returns a HTTP 404 error.
func NotFoundParameterLocale(c *gin.Context, userLocale, parameterName, messageKey string) {
	t := localization.Translate(userLocale)
	NotFoundParameterValue(c, parameterName,
		t(messageKey),
		t("error.request.parameter.locale.not_found.reason", localization.H{"ParameterName": parameterName}),
		t("error.request.parameter.locale.not_found.recovery", localization.H{"ParameterName": parameterName}),
	)
}

// UnauthorizedWithInvalidAuthorizationToken : returns a HTTP 401 error.
func UnauthorizedWithInvalidAuthorizationToken(c *gin.Context, locale string) {
	t := localization.Translate(locale)
	UnauthorizedWithError(c, response.Error{
		Domain:   auth.Domain,
		Code:     auth.InvalidAuthorizationToken,
		Message:  t("error.unauthorized.message"),
		Reason:   t("error.unauthorized.invalid_token.reason"),
		Recovery: t("error.unauthorized.invalid_token.recovery"),
	})
}

// UnauthorizedWithRevokedAuthorizationToken : returns a HTTP 401 error.
func UnauthorizedWithRevokedAuthorizationToken(c *gin.Context, locale string) {
	t := localization.Translate(locale)
	UnauthorizedWithError(c, response.Error{
		Domain:   auth.Domain,
		Code:     auth.RevokedAuthorizationToken,
		Message:  t("error.unauthorized.message"),
		Reason:   t("error.unauthorized.revoked_token.reason"),
		Recovery: t("error.unauthorized.revoked_token.recovery"),
	})
}

// UnauthorizedWithExpiredAuthorizationToken : returns a HTTP 401 error.
func UnauthorizedWithExpiredAuthorizationToken(c *gin.Context, locale string) {
	t := localization.Translate(locale)
	UnauthorizedWithError(c, response.Error{
		Domain:   auth.Domain,
		Code:     auth.ExpiredAuthorizationToken,
		Message:  t("error.unauthorized.message"),
		Reason:   t("error.unauthorized.expired_token.reason"),
		Recovery: t("error.unauthorized.expired_token.recovery"),
	})
}

// UnauthorizedWithInsufficientRights : returns a HTTP 401 error.
func UnauthorizedWithInsufficientRights(c *gin.Context, locale string) {
	t := localization.Translate(locale)
	UnauthorizedWithError(c, response.Error{
		Domain:   auth.Domain,
		Code:     auth.InsufficientRights,
		Message:  t("error.unauthorized.message"),
		Reason:   t("error.unauthorized.insufficient_rights.reason"),
		Recovery: t("error.unauthorized.insufficient_rights.recovery"),
	})
}

// InvalidRequestParameter : returns a HTTP 400 error.
func InvalidRequestParameter(c *gin.Context, locale string) {
	t := localization.Translate(locale)
	InvalidRequestParameterWithDetails(c,
		t("error.request.invalid.parameter.message"),
		t("error.request.invalid.parameter.reason"),
		t("error.request.invalid.parameter.recovery"))
}

// InvalidRequestParameterWithDetails : returns a HTTP 400 error.
func InvalidRequestParameterWithDetails(c *gin.Context, message, reason, recovery string) {
	BadRequestWithError(c, response.Error{
		Domain:   request.Domain,
		Code:     request.InvalidRequest,
		Message:  message,
		Reason:   reason,
		Recovery: recovery,
	})
}

// InvalidParameterValue : returns a HTTP 400 error.
func InvalidParameterValue(c *gin.Context, parameterName, message, reason, recovery string) {
	BadRequestWithError(c, response.Error{
		Domain:   request.Domain,
		Code:     request.InvalidParameterValue,
		Field:    parameterName,
		Message:  message,
		Reason:   reason,
		Recovery: recovery,
	})
}

// NotFoundParameterValue : returns a HTTP 400 error.
func NotFoundParameterValue(c *gin.Context, parameterName, message, reason, recovery string) {
	BadRequestWithError(c, response.Error{
		Domain:   request.Domain,
		Code:     request.NotFoundParameterValue,
		Field:    parameterName,
		Message:  message,
		Reason:   reason,
		Recovery: recovery,
	})
}

// UnableToReadFile : returns a HTTP 500 error.
func UnableToReadFile(c *gin.Context, locale, messageKey string) {
	t := localization.Translate(locale)
	InternalServerErrorWithError(c, response.Error{
		Domain:   file.Domain,
		Code:     file.CanNotReadFile,
		Message:  t(messageKey),
		Reason:   t("error.file.read.reason"),
		Recovery: t("error.file.read.recovery"),
	})
}

// UnableToSaveFile : returns a HTTP 500 error.
func UnableToSaveFile(c *gin.Context, locale, messageKey string) {
	t := localization.Translate(locale)
	InternalServerErrorWithError(c, response.Error{
		Domain:   file.Domain,
		Code:     file.CanNotSaveFile,
		Message:  t(messageKey),
		Reason:   t("error.file.save.reason"),
		Recovery: t("error.file.save.recovery"),
	})
}

// FileNotFound : returns a HTTP 500 error.
func FileNotFound(c *gin.Context, locale, messageKey string) {
	t := localization.Translate(locale)
	InternalServerErrorWithError(c, response.Error{
		Domain:   file.Domain,
		Code:     file.FileNotFound,
		Message:  t(messageKey),
		Reason:   t("error.file.not_found.reason"),
		Recovery: t("error.file.not_found.recovery"),
	})
}

// MaxFileSizeExceeded : returns a HTTP 500 error.
func MaxFileSizeExceeded(c *gin.Context, locale, messageKey string, maxSize uint64) {
	t := localization.Translate(locale)
	PreconditionFailedWithError(c, response.Error{
		Domain:   file.Domain,
		Code:     file.MaxFileSizeExceeded,
		Message:  t(messageKey),
		Reason:   t("error.file.max_file_size_exceeded.reason"),
		Recovery: t("error.file.max_file_size_exceeded.recovery", localization.H{"Size": humanize.Bytes(maxSize)}),
	})
}

// InvalidMimeType : returns a HTTP 500 error.
func InvalidMimeType(c *gin.Context, locale, messageKey string) {
	t := localization.Translate(locale)
	PreconditionFailedWithError(c, response.Error{
		Domain:   file.Domain,
		Code:     file.InvalidMimeType,
		Message:  t(messageKey),
		Reason:   t("error.file.invalid_mime_type.reason"),
		Recovery: t("error.file.invalid_mime_type.recovery"),
	})
}

// UnableToUpdateDatabase : returns a HTTP 500 error.
func UnableToUpdateDatabase(c *gin.Context, locale, messageKey string) {
	t := localization.Translate(locale)
	InternalServerErrorWithError(c, response.Error{
		Domain:   database.Domain,
		Code:     database.CanNotUpdate,
		Message:  t(messageKey),
		Reason:   t("error.database.update.reason"),
		Recovery: t("error.database.update.recovery"),
	})
}

// SendError sends response with the error.
func SendError(c *gin.Context, err error) {
	if errors.IsNotFound(err) {
		NotFound(c)
		return
	}

	InternalServerError(c)
}
