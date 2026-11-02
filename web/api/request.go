package api

// Request is a request to the API.
type Request struct {
	Method                     string // One of the HTTP methods of this package.
	Endpoint                   string // The path of the API, after its version.
	ContentType                string
	Authenticated              bool // If true, the authentication cookie of the incoming request is sent as the auth token.
	Body                       any
	Response                   any // A pointer, filled with the body of the successful responses.
	UseDefaultResponseHandlers bool
	ResponseHandlers           []ResponseHandlerFunc // Called, after the default ones, until one handles the response.
}

// NewRequest returns a GET request which uses the default response handlers.
func NewRequest() (request Request) {
	request = Request{
		Method:                     GET,
		Authenticated:              false,
		UseDefaultResponseHandlers: true,
		ResponseHandlers:           []ResponseHandlerFunc{},
	}

	return
}
