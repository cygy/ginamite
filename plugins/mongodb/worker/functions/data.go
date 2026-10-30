package functions

import (
	"github.com/cygy/ginamite/common/mongo/session"
	"github.com/cygy/ginamite/common/mongo/store"
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
func CreateDefaultDocuments(environment, version string, f func(string, string, store.Database)) {
	db := session.Database()
	f(environment, version, db)
}

// RemoveDefaultDocuments : removes some unused documents from the database.
func RemoveDefaultDocuments(environment, version string, f func(string, string, store.Database)) {
	db := session.Database()
	f(environment, version, db)
}

// UpdateDefaultDocuments : updates some default documents from the database.
func UpdateDefaultDocuments(environment, version string, f func(string, string, store.Database)) {
	db := session.Database()
	f(environment, version, db)
}
