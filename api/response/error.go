package response

import (
	"net/http"

	"github.com/cygy/ginamite/common/log"
	"github.com/cygy/ginamite/common/response"

	"github.com/gin-gonic/gin"
)

// NotFound : returns a HTTP 404 error.
func NotFound(c *gin.Context) {
	c.AbortWithStatus(http.StatusNotFound)
}

// NotFoundWithError : returns a detailed HTTP 404 error.
func NotFoundWithError(c *gin.Context, err response.Error) {
	c.AbortWithStatusJSON(http.StatusNotFound, response.NewError(err))
}

// InternalServerError : returns a HTTP 5OO error.
func InternalServerError(c *gin.Context) {
	c.AbortWithStatus(http.StatusInternalServerError)
}

// InternalServerErrorWithError : returns a detailed HTTP 500 error.
func InternalServerErrorWithError(c *gin.Context, err response.Error) {
	c.Set(log.ErrorContextKey, err.Error())
	c.AbortWithStatusJSON(http.StatusInternalServerError, response.NewError(err))
}

// BadRequest : returns a HTTP 400 error.
func BadRequest(c *gin.Context) {
	c.AbortWithStatus(http.StatusBadRequest)
}

// BadRequestWithError : returns a detailed HTTP 400 error.
func BadRequestWithError(c *gin.Context, err response.Error) {
	c.AbortWithStatusJSON(http.StatusBadRequest, response.NewError(err))
}

// Unauthorized : returns a HTTP 401 error.
func Unauthorized(c *gin.Context) {
	c.AbortWithStatus(http.StatusUnauthorized)
}

// UnauthorizedWithError : returns a detailed HTTP 401 error.
func UnauthorizedWithError(c *gin.Context, err response.Error) {
	c.AbortWithStatusJSON(http.StatusUnauthorized, response.NewError(err))
}

// Forbidden : returns a HTTP 403 error.
func Forbidden(c *gin.Context) {
	c.AbortWithStatus(http.StatusForbidden)
}

// ForbiddenWithError : returns a detailed HTTP 403 error.
func ForbiddenWithError(c *gin.Context, err response.Error) {
	c.AbortWithStatusJSON(http.StatusForbidden, response.NewError(err))
}

// PreconditionFailed : returns a HTTP 412 error.
func PreconditionFailed(c *gin.Context) {
	c.AbortWithStatus(http.StatusPreconditionFailed)
}

// PreconditionFailedWithError : returns a detailed HTTP 412 error.
func PreconditionFailedWithError(c *gin.Context, err response.Error) {
	c.AbortWithStatusJSON(http.StatusPreconditionFailed, response.NewError(err))
}
