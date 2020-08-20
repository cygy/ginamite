package response

import (
	"net/http"

	"github.com/cygy/ginamite/common/response"

	"github.com/gin-gonic/gin"
)

// Ok : returns a HTTP 200 response.
func Ok(c *gin.Context, i interface{}) {
	c.AbortWithStatusJSON(http.StatusOK, i)
}

// OkWithStatus : returns a detailed HTTP 200 response.
func OkWithStatus(c *gin.Context, status string) {
	c.AbortWithStatusJSON(http.StatusOK, response.NewStatus(status))
}

// Created : returns a HTTP 201 response with a JSON object.
func Created(c *gin.Context, i interface{}) {
	c.AbortWithStatusJSON(http.StatusCreated, i)
}

// CreatedWithStatus : returns a HTTP 201 response with a JSON object.
func CreatedWithStatus(c *gin.Context, status string) {
	c.AbortWithStatusJSON(http.StatusCreated, response.NewStatus(status))
}

// CreatedWithStatusAndID : returns a HTTP 201 response with a JSON object.
func CreatedWithStatusAndID(c *gin.Context, status, ID string) {
	c.AbortWithStatusJSON(http.StatusCreated, response.NewStatusAndID(status, ID))
}

// NoContent : returns a HTTP 204 response.
func NoContent(c *gin.Context) {
	c.AbortWithStatus(http.StatusNoContent)
}

// AddRenewedAuthToken : adds a header to the response.
func AddRenewedAuthToken(c *gin.Context, token string) {
	c.Writer.Header().Set(response.RenewedAuthTokenHeader, token)
}
