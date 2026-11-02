package log

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"log/slog"
	"net/http"
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
// It generates a request ID if the request has none. Passwords and the local part of
// the email addresses are obfuscated, the Authorization, Proxy-Authorization and Cookie
// headers are masked, and multipart bodies are not logged. The
// response body is logged, gzip-decoded, only if logResponseBody is true.
func InjectRequestLogger(logResponseBody bool) gin.HandlerFunc {
	// Here are the obfuscators.
	sensitiveObfuscatorRegExp := regexp.MustCompile(`"(password)":"([^\"]*)"`)
	sensitiveObfuscatorReplacingString := []byte(fmt.Sprintf(`"$1":"%s"`, obfuscatorReplacingString))
	emailObfuscatorRegExp := regexp.MustCompile(`([^\/\"]*)@`)
	emailObfuscatorReplacingString := []byte(fmt.Sprintf(`%s@`, obfuscatorReplacingString))

	// Return the middleware.
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
			requestPath = requestPath + "?" + requestQuery
		}

		requestBody := ""
		if strings.HasPrefix(c.Request.Header.Get("Content-Type"), "multipart/form-data") {
			requestBody = "not logged"
		} else if c.Request.Body != nil {
			if body, _ := io.ReadAll(c.Request.Body); body != nil {
				obfuscatedBody := sensitiveObfuscatorRegExp.ReplaceAll(body, sensitiveObfuscatorReplacingString)
				obfuscatedBody = emailObfuscatorRegExp.ReplaceAll(obfuscatedBody, emailObfuscatorReplacingString)
				requestBody = string(obfuscatedBody)

				// Restore the io.ReadCloser to its original state
				c.Request.Body = io.NopCloser(bytes.NewBuffer(body))
			}
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

			obfuscatedResponseBody = emailObfuscatorRegExp.ReplaceAll(obfuscatedResponseBody, emailObfuscatorReplacingString)

			responseLog = append(responseLog, "body", string(obfuscatedResponseBody))
		}

		Info("request", append(attrs, slog.Group("response", responseLog...))...)
	}
}
