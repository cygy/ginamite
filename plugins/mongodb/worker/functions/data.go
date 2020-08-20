package functions

import (
	"github.com/cygy/ginamite/common/mongo/session"
	"github.com/globalsign/mgo"
)

// CreateAllDefaultDocuments : populates the database with the initial documents.
func CreateAllDefaultDocuments(environment, version string) {
}

// RemoveUnusedDefaultDocuments : removes some unused documents from the database.
func RemoveUnusedDefaultDocuments(environment, version string) {
}

// UpdateAllDefaultDocuments : updates some default documents from the database.
func UpdateAllDefaultDocuments(environment, version string) {
}

// CreateDefaultDocuments : populates the database with the initial documents.
func CreateDefaultDocuments(environment, version string, f func(string, string, *mgo.Database)) {
	session, db := session.Copy()
	defer session.Close()

	f(environment, version, db)
}

// RemoveDefaultDocuments : removes some unused documents from the database.
func RemoveDefaultDocuments(environment, version string, f func(string, string, *mgo.Database)) {
	session, db := session.Copy()
	defer session.Close()

	f(environment, version, db)
}

// UpdateDefaultDocuments : updates some default documents from the database.
func UpdateDefaultDocuments(environment, version string, f func(string, string, *mgo.Database)) {
	session, db := session.Copy()
	defer session.Close()

	f(environment, version, db)
}
