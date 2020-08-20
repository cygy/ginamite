package api

import "github.com/cygy/ginamite/common/response"

// Result : result of an API call.
type Result struct {
	StatusCode   int
	RequestError error
	APIError     response.Error
	Handled      bool
}
