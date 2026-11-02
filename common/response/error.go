// Package response defines the JSON structures returned by the API.
package response

import "fmt"

// Error is an error with more details, serialized in the responses.
type Error struct {
	Code     int    `json:"code"`
	Domain   string `json:"domain"`
	Message  string `json:"message,omitempty"`
	Reason   string `json:"reason,omitempty"`
	Recovery string `json:"recovery,omitempty"`
	Field    string `json:"field,omitempty"`
}

// Error returns the message, the reason and the recovery of the error, separated by spaces.
func (e *Error) Error() string {
	return fmt.Sprintf("%s %s %s", e.Message, e.Reason, e.Recovery)
}

// IsNil reports whether the error is empty, i.e. has neither code nor domain.
func (e *Error) IsNil() bool {
	return e.Code == 0 && len(e.Domain) == 0
}
