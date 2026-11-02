package validator

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/cygy/ginamite/common/log"

	resty "github.com/go-resty/resty/v2"
	"github.com/sirupsen/logrus"
)

func apiClient(host string, timeOut, retryCount int, debugMode bool) *resty.Client {
	client := resty.New()
	client.SetBaseURL(host).
		SetHeader("Accept", "application/json").
		SetTimeout(time.Duration(timeOut) * time.Millisecond).
		SetRetryCount(retryCount).
		SetDebug(debugMode)

	return client
}

func extractClaimsFromJWTToken(token string, value any, validator TokenValidator) bool {
	segments := strings.Split(token, ".")
	if len(segments) != 3 {
		log.WithFields(logrus.Fields{
			"source": validator.SourceName(),
		}).Error("wrong JWT token format")
		return false
	}

	parsed, err := base64.RawURLEncoding.DecodeString(segments[1])
	if err != nil {
		log.WithFields(logrus.Fields{
			"source": validator.SourceName(),
			"error":  err.Error(),
		}).Error("unable to parse JWT token")
		return false
	}

	if !logParseToken(validator, parsed, value) {
		return false
	}

	return true
}

func logResponseToken(validator TokenValidator, r *resty.Response, err error) bool {
	if err != nil || r.StatusCode() != http.StatusOK {
		entry := log.WithFields(logrus.Fields{
			"source":     validator.SourceName(),
			"statusCode": r.StatusCode(),
		})
		if err != nil {
			entry = entry.WithField("error", err.Error())
		}
		entry.Error("unable to validate token")

		return false
	}

	return true
}

func logParseToken(validator TokenValidator, rawToken []byte, value any) bool {
	if err := json.Unmarshal(rawToken, value); err != nil {
		log.WithFields(logrus.Fields{
			"source": validator.SourceName(),
			"error":  err.Error(),
		}).Error("unable to parse token")

		return false
	}

	return true
}
