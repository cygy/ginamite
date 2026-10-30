package api

import "github.com/cygy/ginamite/common/response"

// Result is the result of an API call.
type Result struct {
	StatusCode   int // Zero if the API did not respond.
	RequestError error
	APIError     response.Error // Nil, see response.Error.IsNil, if the API returned no error.
	Handled      bool           // True if a response handler handled the response: nothing is left to do.
}
