package terms

import (
	"fmt"

	"github.com/cygy/ginamite/api/context"
	"github.com/cygy/ginamite/api/request"
	"github.com/cygy/ginamite/api/response"
	"github.com/cygy/ginamite/common/config"
	"github.com/cygy/ginamite/common/localization"
	"github.com/cygy/ginamite/common/log"
	r "github.com/cygy/ginamite/common/response"

	"io/ioutil"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// GetTerms returns the current terms and version.
func GetTerms(c *gin.Context) {
	locale := context.GetLocale(c)
	version := config.Main.Terms.Version
	filename := fmt.Sprintf("%s/%s", version, locale)

	terms, ok := termsContents[filename]
	if !ok {
		filePath := fmt.Sprintf("%s/%s.txt", config.Main.Terms.Path, filename)
		bytes, err := ioutil.ReadFile(filePath)
		if err != nil {
			log.WithFields(logrus.Fields{
				"version": version,
				"locale":  locale,
				"path":    filePath,
				"error":   err.Error(),
			}).Error("unable to read the terms")
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
		"version": config.Main.Terms.Version,
	})
}

// AcceptTerms validates and acccepts the current terms.
func AcceptTerms(c *gin.Context) {
	var jsonBody struct {
		Version string `json:"version"`
	}
	request.DecodeBody(c, &jsonBody)

	locale := context.GetLocale(c)
	t := localization.Translate(locale)

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
	if termsVersion != config.Main.Terms.Version {
		message := t("error.terms.version.invalid.message")
		reason := t("error.terms.version.invalid.reason")
		recovery := t("error.terms.version.invalid.recovery")

		response.InvalidParameterValue(c, "version", message, reason, recovery)
		return
	}

	if UpdateVersionOfTermsAcceptedByUser != nil {
		userID := context.GetUserID(c)
		if err := UpdateVersionOfTermsAcceptedByUser(c, userID, termsVersion); err != nil {
			response.InternalServerError(c)
			return
		}
	}

	response.Ok(c, r.NewStatus(t("status.terms.accepted")))
}
