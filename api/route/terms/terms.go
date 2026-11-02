// Package terms provides the handlers about the terms and their acceptation.
package terms

import (
	"fmt"
	"os"
	"sync"

	"github.com/cygy/ginamite/common/ctxkey"

	ctxLocale "github.com/cygy/ginamite/api/context/locale"
	ctxUser "github.com/cygy/ginamite/api/context/user"
	"github.com/cygy/ginamite/api/request"
	"github.com/cygy/ginamite/api/response"
	"github.com/cygy/ginamite/common/log"
	r "github.com/cygy/ginamite/common/response"

	"github.com/gin-gonic/gin"
)

// GetTerms returns the handler sending the current terms and version. The contents of the terms are cached by version
// and locale in the handler.
func GetTerms() gin.HandlerFunc {
	var mutex sync.Mutex
	termsContents := map[string]string{} // The localized contents of the terms.

	return func(c *gin.Context) {
		mutex.Lock()
		defer mutex.Unlock()

		getTerms(c, termsContents)
	}
}

// getTerms sends the terms, read from the file if they are not in the cache.
func getTerms(c *gin.Context, termsContents map[string]string) {
	locale := ctxLocale.Get(c)
	version := ctxkey.Config.Get(c).Terms.Version
	filename := fmt.Sprintf("%s/%s", version, locale)

	terms, ok := termsContents[filename]
	if !ok {
		filePath := fmt.Sprintf("%s/%s.txt", ctxkey.Config.Get(c).Terms.Path, filename)
		bytes, err := os.ReadFile(filePath)
		if err != nil {
			log.Error("unable to read the terms",
				"version", version,
				"locale", locale,
				"path", filePath,
				"error", err.Error(),
			)
			response.InternalServerError(c)
			return
		}

		terms = string(bytes)
		termsContents[filename] = terms
	}

	response.Ok(c, response.H{
		"terms":   terms,
		"version": version,
	})
}

// GetTermsVersion returns the current version of the terms.
func GetTermsVersion(c *gin.Context) {
	response.Ok(c, response.H{
		"version": ctxkey.Config.Get(c).Terms.Version,
	})
}

// AcceptTerms validates and acccepts the current terms.
func AcceptTerms(c *gin.Context) {
	var jsonBody struct {
		Version string `json:"version"`
	}
	request.DecodeBody(c, &jsonBody)

	locale := ctxLocale.Get(c)
	t := ctxkey.Localization.Get(c).Translate(locale)

	// Terms version is mandatory.
	termsVersion := jsonBody.Version
	if len(termsVersion) == 0 {
		message := t("error.registration.message")
		reason := t("error.registration.terms_version.not_found.reason")
		recovery := t("error.registration.terms_version.not_found.recovery")

		response.NotFoundParameterValue(c, "version", message, reason, recovery)
		return
	}

	// The latest version of the terms must be accepted.
	if termsVersion != ctxkey.Config.Get(c).Terms.Version {
		message := t("error.terms.version.invalid.message")
		reason := t("error.terms.version.invalid.reason")
		recovery := t("error.terms.version.invalid.recovery")

		response.InvalidParameterValue(c, "version", message, reason, recovery)
		return
	}

	if storage := StorageOf(c); storage != nil {
		userID := ctxUser.GetID(c)
		if err := storage.UpdateVersionOfTermsAcceptedByUser(c, userID, termsVersion); err != nil {
			response.InternalServerError(c)
			return
		}
	}

	response.Ok(c, r.NewStatus(t("status.terms.accepted")))
}
