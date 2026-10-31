package kafka

import "github.com/cygy/ginamite/common/log"

// Recover recovers from a panic and logs it as a warning.
// Defer it in the message consumers so that a panic does not stop them.
func Recover() {
	if r := recover(); r != nil {
		log.WithField("panic", r).Warning("recover from panic")
	}
}
