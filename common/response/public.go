package response

// NewError : returns a Response struct.
func NewError(err Error) Response {
	return Response{
		Error: &err,
	}
}

// NewStatus : returns a new Response struct.
func NewStatus(status string) Response {
	return Response{
		Status: status,
	}
}

// NewStatusAndID : returns a new Response struct.
func NewStatusAndID(status, ID string) Response {
	return Response{
		Status: status,
		ID:     ID,
	}
}
