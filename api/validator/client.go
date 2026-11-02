package validator

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/cygy/ginamite/common/log"

	resty "github.com/go-resty/resty/v2"
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
		log.Error("wrong JWT token format", "source", validator.SourceName())
		return false
	}

	parsed, err := base64.RawURLEncoding.DecodeString(segments[1])
	if err != nil {
		log.Error("unable to parse JWT token",
			"source", validator.SourceName(),
			"error", err.Error(),
		)
		return false
	}

	if !logParseToken(validator, parsed, value) {
		return false
	}

	return true
}

func logResponseToken(validator TokenValidator, r *resty.Response, err error) bool {
	if err != nil || r.StatusCode() != http.StatusOK {
		attrs := []any{"source", validator.SourceName(), "statusCode", r.StatusCode()}
		if err != nil {
			attrs = append(attrs, "error", err.Error())
		}
		log.Error("unable to validate token", attrs...)

		return false
	}

	return true
}

func logParseToken(validator TokenValidator, rawToken []byte, value any) bool {
	if err := json.Unmarshal(rawToken, value); err != nil {
		log.Error("unable to parse token",
			"source", validator.SourceName(),
			"error", err.Error(),
		)

		return false
	}

	return true
}
