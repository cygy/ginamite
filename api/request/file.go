package request

import (
	"mime/multipart"
	"net/http"
	"strings"

	ctxLocale "github.com/cygy/ginamite/api/context/locale"
	"github.com/cygy/ginamite/api/response"
	"github.com/cygy/ginamite/common/log"

	"github.com/gin-gonic/gin"
)

// GetUploadedFile returns the uploaded file.
func GetUploadedFile(c *gin.Context, fieldName string, maxSize int64, checkIsImage bool, logContext, errorMessageKey string) *multipart.FileHeader {
	locale := ctxLocale.Get(c)

	file, err := c.FormFile(fieldName)
	if err != nil {
		log.Error("file not found",
			"context", logContext,
			"error", err.Error(),
		)
		response.FileNotFound(c, locale, errorMessageKey)
		return nil
	}

	if file.Size > maxSize {
		log.Error("max file size exceeded", "context", logContext)
		response.MaxFileSizeExceeded(c, locale, errorMessageKey, uint64(maxSize)) //nolint:gosec // maxSize is a positive limit set by the code
		return nil
	}

	if checkIsImage {
		f, err := file.Open()
		if err != nil {
			log.Error("unable to open the file",
				"context", logContext,
				"error", err.Error(),
			)
			response.UnableToReadFile(c, locale, errorMessageKey)
			return nil
		}

		defer func() {
			if err := f.Close(); err != nil {
				log.Error("unable to close the uploaded file", "error", err)
			}
		}()

		buff := make([]byte, 512) // docs tell that it take only first 512 bytes into consideration
		if _, err = f.Read(buff); err != nil {
			log.Error("unable to read the file",
				"context", logContext,
				"error", err.Error(),
			)
			response.UnableToReadFile(c, locale, errorMessageKey)
			return nil
		}

		mimeType := http.DetectContentType(buff)
		if !strings.HasPrefix(mimeType, "image/") {
			log.Error("the file is not an image",
				"context", logContext,
				"error", "invalid mime-type",
				"mime-type", mimeType,
			)
			response.InvalidMimeType(c, locale, errorMessageKey)
			return nil
		}
	}

	return file
}

// SaveUploadedFile saves the uploaded file to the destination.
func SaveUploadedFile(c *gin.Context, file *multipart.FileHeader, destination, logContext, errorMessageKey string) bool {
	if err := c.SaveUploadedFile(file, destination); err != nil {
		log.Error("unable to save the file",
			"context", logContext,
			"error", err.Error(),
		)

		locale := ctxLocale.Get(c)
		response.UnableToSaveFile(c, locale, errorMessageKey)

		return false
	}

	return true
}
