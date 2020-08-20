package response

import "fmt"

// Error : error with more details.
type Error struct {
	Code     int    `json:"code"`
	Domain   string `json:"domain"`
	Message  string `json:"message,omitempty"`
	Reason   string `json:"reason,omitempty"`
	Recovery string `json:"recovery,omitempty"`
	Field    string `json:"field,omitempty"`
}

// Error : returns a string.
func (e *Error) Error() string {
	return fmt.Sprintf("%s %s %s", e.Message, e.Reason, e.Recovery)
}

// IsNil : returns true if it is an empty struct.
func (e *Error) IsNil() bool {
	return e.Code == 0 && len(e.Domain) == 0
}
