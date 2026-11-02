package log

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	obfuscatorReplacingString = "***"

	// ErrorContextKey is the key of the error in the Gin context. If set, the error is logged with the request.
	ErrorContextKey = "request_error"

	// RequestIDHeaderName is the name of the header containing the request ID.
	RequestIDHeaderName = "X-Request-ID"
)

// sensitiveHeaders are the request headers whose value is never logged.
var sensitiveHeaders = map[string]bool{
	"Authorization":       true,
	"Proxy-Authorization": true,
	"Cookie":              true,
}

var (
	// pairRegExp matches the JSON pairs whose value is a string: the key with its colon, then the value. The escaped
	// characters, such as \", are part of the strings.
	pairRegExp = regexp.MustCompile(`("(?:[^"\\]|\\.)*"\s*:\s*)"(?:[^"\\]|\\.)*"`)
	// sensitiveKeyRegExp matches the names of the JSON keys and of the query parameters holding a secret, such as
	// "password", "new_password", "auth_token" or "client_secret".
	sensitiveKeyRegExp = regexp.MustCompile(`(?i)password|token|secret`)
	// emailRegExp matches the local part of the email addresses.
	emailRegExp = regexp.MustCompile(`([^\/\"]*)@`)
)

// obfuscateBody masks the secrets and the local part of the email addresses of a body.
func obfuscateBody(body []byte) []byte {
	body = pairRegExp.ReplaceAllFunc(body, func(pair []byte) []byte {
		key := pairRegExp.FindSubmatch(pair)[1]
		if !sensitiveKeyRegExp.Match(key) {
			return pair
		}

		return append(key, `"`+obfuscatorReplacingString+`"`...)
	})
	return emailRegExp.ReplaceAll(body, []byte(obfuscatorReplacingString+"@"))
}

// obfuscateQuery masks the values of the query parameters holding a secret. A query which can not be parsed is
// masked entirely, since its secrets can not be found.
func obfuscateQuery(rawQuery string) string {
	values, err := url.ParseQuery(rawQuery)
	if err != nil {
		return obfuscatorReplacingString
	}

	for key := range values {
		if sensitiveKeyRegExp.MatchString(key) {
			values[key] = []string{obfuscatorReplacingString}
		}
	}

	return values.Encode()
}

// readCloser reads from a reader and closes a closer.
type readCloser struct {
	io.Reader
	io.Closer
}

// requestLogWriter copies the response body while writing it.
type requestLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

// Write writes b to the response and keeps a copy of it.
func (w requestLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// InjectRequestLogger returns a middleware logging each request and its response.
// It generates a request ID if the request has none. The values of the JSON keys and of the
// query parameters containing password, token or secret, and the local part of the email
// addresses, are obfuscated in the bodies and the query; the Authorization,
// Proxy-Authorization and Cookie headers are masked, and multipart bodies are not logged.
// The response body is logged, gzip-decoded, only if logResponseBody is true.
func InjectRequestLogger(logResponseBody bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the request details
		if len(c.Request.Header.Get(RequestIDHeaderName)) == 0 {
			c.Request.Header.Set(RequestIDHeaderName, uuid.NewString())
		}

		requestIP := c.ClientIP()

		requestMethod := c.Request.Method
		requestPath := c.Request.URL.Path
		requestQuery := c.Request.URL.RawQuery
		if len(requestQuery) > 0 {
			requestPath = requestPath + "?" + obfuscateQuery(requestQuery)
		}

		requestBody := ""
		if strings.HasPrefix(c.Request.Header.Get("Content-Type"), "multipart/form-data") {
			requestBody = "not logged"
		} else if c.Request.Body != nil {
			// A body read partially, such as one exceeding the limit of a http.MaxBytesReader, is logged as read, and
			// the handler reads it in full: the bytes read, then the rest of the original body and its error.
			original := c.Request.Body
			body, _ := io.ReadAll(original)
			requestBody = string(obfuscateBody(body))
			c.Request.Body = readCloser{Reader: io.MultiReader(bytes.NewReader(body), original), Closer: original}
		}

		requestHeaders := map[string]string{}
		for key, value := range c.Request.Header {
			if len(value) == 0 {
				continue
			}

			if sensitiveHeaders[http.CanonicalHeaderKey(key)] {
				requestHeaders[key] = obfuscatorReplacingString
			} else {
				requestHeaders[key] = value[0]
			}
		}

		// Replace the writer
		var writer *requestLogWriter
		if logResponseBody {
			writer = &requestLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
			c.Writer = writer
		}

		// Start timer
		start := time.Now()

		// Process request
		c.Next()

		// Stop timer
		end := time.Now()
		latency := end.Sub(start)

		attrs := []any{
			"IP", requestIP,
			"latency", latency,
			slog.Group("request",
				"method", requestMethod,
				"URI", requestPath,
				"headers", requestHeaders,
				"body", requestBody,
			),
		}

		// Get the error if provided
		if err, errExists := c.Get(ErrorContextKey); errExists {
			attrs = append(attrs, "error", fmt.Sprint(err))
		}

		// Get the response details
		responseLog := []any{"code", c.Writer.Status()}

		if logResponseBody {
			obfuscatedResponseBody := writer.body.Bytes()

			// Check the first two bytes to know if a bytes array is gzipped or not.
			bufferReader := bufio.NewReader(bytes.NewReader(obfuscatedResponseBody))
			if testBytes, err := bufferReader.Peek(2); err == nil && testBytes[0] == 31 && testBytes[1] == 139 {
				gzipReader, err := gzip.NewReader(bytes.NewBuffer(obfuscatedResponseBody))
				if err == nil {
					var data []byte
					if data, err = io.ReadAll(gzipReader); err == nil {
						obfuscatedResponseBody = data
					}

					if closeErr := gzipReader.Close(); err == nil {
						err = closeErr
					}
				}

				if err != nil {
					obfuscatedResponseBody = []byte{}
					responseLog = append(responseLog, "error", "can not uncompress body")
				}
			}

			obfuscatedResponseBody = obfuscateBody(obfuscatedResponseBody)

			responseLog = append(responseLog, "body", string(obfuscatedResponseBody))
		}

		Info("request", append(attrs, slog.Group("response", responseLog...))...)
	}
}
