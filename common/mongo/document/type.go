package document

import "github.com/globalsign/mgo"

// Document : interface for a MongoDB document.
type Document interface {
	IsSaved() bool
	Save(db *mgo.Database) (string, error)
	Update(update interface{}, db *mgo.Database) error
	Delete(db *mgo.Database) error
	GetByID(id string, db *mgo.Database) error
	SetID(id string)
	GetID() string
}
