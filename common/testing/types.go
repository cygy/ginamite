package testing

import (
	"net/http"
	"testing"

	"github.com/cygy/ginamite/common/response"
)

// H shorthand to a map struct
type H map[string]any

// S shorthand to a map struct
type S map[string]string

// Request struct representing a request to test
type Request struct {
	T                *testing.T
	Method           string
	Endpoint         string
	Parameter        string
	Body             H
	AuthToken        string
	RenewedAuthToken string
	Headers          S
	ExpectedResponse struct {
		Code             int
		Status           string
		Error            response.Error
		RenewedAuthToken bool
	}
	Response any
}

// NewRequest : returns a new struct 'Request'
func NewRequest(t *testing.T) Request {
	request := Request{
		T:      t,
		Method: GET,
	}

	request.ExpectedResponse.Code = http.StatusOK

	return request
}
