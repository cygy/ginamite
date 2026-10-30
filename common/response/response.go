package response

// RenewedAuthTokenHeader is the name of the header containing the renewed auth token.
const RenewedAuthTokenHeader = "X-Renewed-Auth-Token"

// Response is the structured JSON object returned by the API.
type Response struct {
	Error  *Error `json:"error,omitempty"`
	Status string `json:"status,omitempty"`
	ID     string `json:"id,omitempty"`
}

// IsErrorNil reports whether the response has no error, or an empty one.
func (r *Response) IsErrorNil() bool {
	return r.Error == nil || r.Error.IsNil()
}

// NewError returns a Response containing the error.
func NewError(err Error) Response {
	return Response{
		Error: &err,
	}
}

// NewStatus returns a Response containing the status.
func NewStatus(status string) Response {
	return Response{
		Status: status,
	}
}

// NewStatusAndID returns a Response containing the status and the ID.
func NewStatusAndID(status, id string) Response {
	return Response{
		Status: status,
		ID:     id,
	}
}
