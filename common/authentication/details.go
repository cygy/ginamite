package authentication

import "time"

// Details holds the details of an authentication.
type Details struct {
	UserID         string
	Device         Device
	Source         string
	Method         string
	IPAddress      string
	Key            string
	CreationDate   time.Time
	ExpirationDate time.Time
}
