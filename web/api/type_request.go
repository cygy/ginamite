package api

// Request : request of an API call.
type Request struct {
	Method                     string
	Endpoint                   string
	ContentType                string
	Authenticated              bool
	Body                       interface{}
	Response                   interface{}
	UseDefaultResponseHandlers bool
	ResponseHandlers           []ResponseHandlerFunc
}

// NewRequest : returns a Request struct.
func NewRequest() (request Request) {
	request = Request{
		Method:                     GET,
		Authenticated:              false,
		UseDefaultResponseHandlers: true,
		ResponseHandlers:           []ResponseHandlerFunc{},
	}

	return
}
