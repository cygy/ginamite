package response

// Response : structured JSON object
type Response struct {
	Error  *Error `json:"error,omitempty"`
	Status string `json:"status,omitempty"`
	ID     string `json:"id,omitempty"`
}

// IsErrorNil : returns true if it is an empty error.
func (r *Response) IsErrorNil() bool {
	return r.Error == nil || r.Error.IsNil()
}
