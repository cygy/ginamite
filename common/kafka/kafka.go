package kafka

import (
	"errors"

	"github.com/cygy/ginamite/common/log"
)

// Recover recovers from a panic and logs it as a warning.
// The consumers do not need it: a panic of an action is recovered and handled like an error.
func Recover() {
	if r := recover(); r != nil {
		log.WithField("panic", r).Warning("recover from panic")
	}
}

// permanentError is an error handling the message again can not fix.
type permanentError struct{ err error }

func (e permanentError) Error() string { return e.err.Error() }

func (e permanentError) Unwrap() error { return e.err }

// Permanent marks the error of an action as one a retry can not fix, such as a message which can not be decoded: the
// message is dead-lettered at once, instead of being tried again. It returns nil if err is nil.
func Permanent(err error) error {
	if err == nil {
		return nil
	}

	return permanentError{err}
}

// IsPermanent reports whether the error, or one it wraps, is marked with Permanent.
func IsPermanent(err error) bool {
	return errors.As(err, &permanentError{})
}
