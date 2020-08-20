package log

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"fmt"
	"io/ioutil"
	"regexp"
	"strings"
	"time"

	"github.com/cygy/ginamite/common/request"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

const (
	obfuscatorReplacingString = "***"

	// ErrorContextKey field name of the error in the context, if provided this error will be logged.
	ErrorContextKey = "request_error"

	// RequestIDHeaderName header name of the request ID value.
	RequestIDHeaderName = "X-Request-ID"
)

type requestLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w requestLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// InjectRequestLogger log the requests and the responses.
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
			uuid := uuid.New()
			c.Request.Header.Set(RequestIDHeaderName, uuid.String())
		}

		requestIP := request.GetRealIPAddress(c)

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
			if body, _ := ioutil.ReadAll(c.Request.Body); body != nil {
				obfuscatedBody := sensitiveObfuscatorRegExp.ReplaceAll(body, sensitiveObfuscatorReplacingString)
				obfuscatedBody = emailObfuscatorRegExp.ReplaceAll(obfuscatedBody, emailObfuscatorReplacingString)
				requestBody = string(obfuscatedBody[:])

				// Restore the io.ReadCloser to its original state
				c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(body))
			}
		}

		requestHeaders := map[string]string{}
		for key, value := range c.Request.Header {
			requestHeaders[key] = value[0]
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

		entry := WithFields(logrus.Fields{
			"IP":      requestIP,
			"latency": latency,
			"request": logrus.Fields{
				"method":  requestMethod,
				"URI":     requestPath,
				"headers": requestHeaders,
				"body":    requestBody,
			},
		})

		// Get the error if provided
		if err, errExists := c.Get(ErrorContextKey); errExists {
			entry = entry.WithFields(logrus.Fields{
				"error": err.(string),
			})
		}

		// Get the response details
		responseLog := logrus.Fields{
			"code": c.Writer.Status(),
		}

		if logResponseBody {
			obfuscatedResponseBody := writer.body.Bytes()

			// Check the first two bytes to know if a bytes array is gzipped or not.
			bufferReader := bufio.NewReader(bytes.NewReader(obfuscatedResponseBody))
			if testBytes, err := bufferReader.Peek(2); err == nil && testBytes[0] == 31 && testBytes[1] == 139 {
				gzipReader, err := gzip.NewReader(bytes.NewBuffer(obfuscatedResponseBody))
				if err == nil {
					var data []byte
					if data, err = ioutil.ReadAll(gzipReader); err == nil {
						obfuscatedResponseBody = data
					}
				}

				gzipReader.Close()

				if err != nil {
					obfuscatedResponseBody = []byte{}
					responseLog["error"] = "can not uncompress body"
				}
			}

			obfuscatedResponseBody = emailObfuscatorRegExp.ReplaceAll(obfuscatedResponseBody, emailObfuscatorReplacingString)

			responseLog["body"] = string(obfuscatedResponseBody[:])
		}

		entry.WithFields(logrus.Fields{
			"response": responseLog,
		}).Info("request")
	}
}
