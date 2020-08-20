package validator

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/cygy/ginamite/common/log"

	"github.com/sirupsen/logrus"
	resty "gopkg.in/resty.v1"
)

// TokenInfos struct about the informations of a token
type TokenInfos struct {
	Token     string
	UserID    string
	Email     string
	LastName  string
	FirstName string
	CreatedAt time.Time
	ExpiresAt time.Time
}

// TokenValidator validates a auth token
type TokenValidator interface {
	SourceName() string
	VerifyTokenAndExtractInfos(token string) (TokenInfos, bool)
}

// Utils
func apiClient(host string, timeOut, retryCount int, debugMode bool) *resty.Client {
	client := resty.New()
	client.SetHostURL(host).
		SetHeader("Accept", "application/json").
		SetTimeout(time.Duration(timeOut) * time.Millisecond).
		SetRetryCount(retryCount).
		SetDebug(debugMode)

	return client
}

func extractClaimsFromJWTToken(token string, value interface{}, validator TokenValidator) bool {
	segments := strings.Split(token, ".")
	if len(segments) != 3 {
		log.WithFields(logrus.Fields{
			"token":  token,
			"source": validator.SourceName(),
		}).Error("wrong JWT token format")
		return false
	}

	parsed, err := base64.StdEncoding.DecodeString(segments[1])
	if err != nil {
		log.WithFields(logrus.Fields{
			"token":  token,
			"source": validator.SourceName(),
			"error":  err.Error(),
		}).Error("unable to parse JWT oken")
		return false
	}

	if !logParseToken(token, validator, parsed, value) {
		return false
	}

	return true
}

func logResponseToken(token string, validator TokenValidator, r *resty.Response, err error) bool {
	if err != nil || r.StatusCode() != http.StatusOK {
		entry := log.WithFields(logrus.Fields{
			"token":      token,
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

func logParseToken(token string, validator TokenValidator, rawToken []byte, value interface{}) bool {
	if err := json.Unmarshal(rawToken, value); err != nil {
		log.WithFields(logrus.Fields{
			"token":  token,
			"source": validator.SourceName(),
			"error":  err.Error(),
		}).Error("unable to parse token")

		return false
	}

	return true
}
