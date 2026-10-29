package session

import "go.mongodb.org/mongo-driver/v2/mongo"

// Session : session of a mongo database.
// The client is safe for concurrent use and manages its own connection pool.
type Session struct {
	Client       *mongo.Client
	DatabaseName string
}

// NewSession : returns a new Session struct
func NewSession() (session *Session) {
	session = &Session{}
	return
}
