package kafka

import "github.com/cygy/ginamite/common/log"

// Recover recovers from a panic and logs it as a warning.
// The consumers do not need it: a panic of an action is recovered and handled like an error.
func Recover() {
	if r := recover(); r != nil {
		log.Warn("recover from panic", "panic", r)
	}
}
