// Package functions provides the functions of the workers working on the MongoDB database of the plugin.
package functions

import (
	"context"

	"github.com/cygy/ginamite/plugins/mongodb/mongo/store"
	"github.com/cygy/ginamite/worker"
)

var (
	_ worker.TokensStorage        = (*Functions)(nil)
	_ worker.AccountsStorage      = (*Functions)(nil)
	_ worker.UsersStorage         = (*Functions)(nil)
	_ worker.NotificationsStorage = (*Functions)(nil)
	_ worker.LocationsStorage     = (*Functions)(nil)
	_ worker.EmailsStorage        = (*Functions)(nil)
)

// UserHook is a function of the application run on the documents referencing a user.
type UserHook func(ctx context.Context, userID string, db store.Database) error

// Hooks are the functions of the application which complete the default ones. They are all optional.
type Hooks struct {
	// SanitizeAccounts updates, or deletes, the data left by the deleted users and returns the count of documents
	// sanitized.
	SanitizeAccounts func(ctx context.Context, db store.Database) uint
	// IsUserEmpty reports whether a user never did any action. Every user is empty if it is nil.
	IsUserEmpty func(ctx context.Context, userID string, db store.Database) bool
	// UpdateDocumentsReferencingToDeletedUser runs when a user is deleted.
	UpdateDocumentsReferencingToDeletedUser UserHook
	// UpdateDocumentsReferencingToDisabledUser runs when a user is disabled.
	UpdateDocumentsReferencingToDisabledUser UserHook
	// UpdateDocumentsReferencingToEnabledUser runs when a user is enabled.
	UpdateDocumentsReferencingToEnabledUser UserHook
	// UpdateDocumentsReferencingToUpdatedUserSocialNetworks runs when the social networks of a user are updated.
	UpdateDocumentsReferencingToUpdatedUserSocialNetworks UserHook
}

// Functions are the functions of the workers working on a database: it implements the storages of the worker, see
// worker.Storage. Create it with New.
type Functions struct {
	// Hooks complete the default functions.
	Hooks Hooks

	database func() store.Database
}

// New returns the functions working on the database returned by the function, which is called each time a function
// needs the database.
func New(database func() store.Database) *Functions {
	return &Functions{database: database}
}
