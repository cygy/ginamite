package request

import (
	"mime/multipart"
	"net/http"
	"strings"

	"github.com/cygy/ginamite/api/context"
	"github.com/cygy/ginamite/api/response"
	"github.com/cygy/ginamite/common/log"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// GetUploadedFile : returns the uploaded file.
func GetUploadedFile(c *gin.Context, fieldName string, maxSize int64, checkIsImage bool, logContext, errorMessageKey string) *multipart.FileHeader {
	locale := context.GetLocale(c)

	file, err := c.FormFile(fieldName)

	if err != nil {
		log.WithFields(logrus.Fields{
			"context": logContext,
			"error":   err.Error(),
		}).Error("file not found")
		response.FileNotFound(c, locale, errorMessageKey)
		return nil
	}

	if file.Size > maxSize {
		log.WithFields(logrus.Fields{
			"context": logContext,
		}).Error("max file size exceeded")
		response.MaxFileSizeExceeded(c, locale, errorMessageKey, uint64(maxSize))
		return nil
	}

	if checkIsImage {
		f, err := file.Open()
		if err != nil {
			log.WithFields(logrus.Fields{
				"context": logContext,
				"error":   err.Error(),
			}).Error("unable to open the file")
			response.UnableToReadFile(c, locale, errorMessageKey)
			return nil
		}

		defer f.Close()

		buff := make([]byte, 512) // docs tell that it take only first 512 bytes into consideration
		if _, err = f.Read(buff); err != nil {
			log.WithFields(logrus.Fields{
				"context": logContext,
				"error":   err.Error(),
			}).Error("unable to read the file")
			response.UnableToReadFile(c, locale, errorMessageKey)
			return nil
		}

		mimeType := http.DetectContentType(buff)
		if !strings.HasPrefix(mimeType, "image/") {
			log.WithFields(logrus.Fields{
				"context":   logContext,
				"error":     "invalid mime-type",
				"mime-type": mimeType,
			}).Error("the file is not an image")
			response.InvalidMimeType(c, locale, errorMessageKey)
			return nil
		}
	}

	return file
}

// SaveUploadedFile : saves the uploaded file to the destination.
func SaveUploadedFile(c *gin.Context, file *multipart.FileHeader, destination, logContext, errorMessageKey string) bool {
	if err := c.SaveUploadedFile(file, destination); err != nil {
		log.WithFields(logrus.Fields{
			"context": logContext,
			"error":   err.Error(),
		}).Error("unable to save the file")

		locale := context.GetLocale(c)
		response.UnableToSaveFile(c, locale, errorMessageKey)

		return false
	}

	return true
}
