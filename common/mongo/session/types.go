package session

import "github.com/globalsign/mgo"

// Session : session of a mongo database.
type Session struct {
	Session      *mgo.Session
	DatabaseName string
}

// NewSession : returns a new Session struct
func NewSession() (session *Session) {
	session = &Session{}
	return
}
