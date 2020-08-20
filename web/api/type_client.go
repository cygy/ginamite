package api

import (
	"fmt"
	"time"

	"github.com/cygy/ginamite/common/response"

	"github.com/gin-gonic/gin"
	resty "gopkg.in/resty.v1"
)

// ResponseHandlerFunc : handler for the API responses.
type ResponseHandlerFunc func(c *gin.Context, statusCode int, requestError error, APIError response.Error) bool

// Client : client of an API.
type Client struct {
	client           *resty.Client
	responseHandlers []ResponseHandlerFunc
}

// NewClient : returns a new API client struct.
func NewClient(host, APIVersion string, port, timeout, retryCount int, debug bool) (client *Client) {
	hostURL := fmt.Sprintf("%s:%d/%s", host, port, APIVersion)

	apiClient := resty.New()
	apiClient.SetHostURL(hostURL).
		SetHeader("Accept", "application/json").
		SetHeader("Accept-Encoding", "gzip").
		SetTimeout(time.Duration(timeout) * time.Millisecond).
		SetRetryCount(retryCount).
		SetDebug(debug)

	client = &Client{
		client:           apiClient,
		responseHandlers: []ResponseHandlerFunc{},
	}
	return
}
