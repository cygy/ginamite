package user

import (
	"context"

	"github.com/gin-gonic/gin"
)

// storageKey is the key of the storage in the Gin context.
const storageKey = "_user_storage"

// Storage is the storage of the users for the administration, provided by the application, such as the one of the
// storage plugin. The context is the one of the request.
type Storage interface {
	// GetEnabledUsersForAdmin returns the list of the users to display in the admin panel.
	GetEnabledUsersForAdmin(ctx context.Context) (any, error)
	// GetDisabledUsersForAdmin returns the list of the disabled users to display in the admin panel.
	GetDisabledUsersForAdmin(ctx context.Context) (any, error)

	// UpdatePasswordByAdmin updates the password of a user.
	UpdatePasswordByAdmin(ctx context.Context, userID, password string) error
	// UpdateInformationsByAdmin updates the informations of a user.
	UpdateInformationsByAdmin(ctx context.Context, userID, firstName, lastName, emailAddress string, validEmailAddress bool) error
	// UpdateDescriptionByAdmin updates the description of a user.
	UpdateDescriptionByAdmin(ctx context.Context, userID, locale, description string) error
	// UpdateSocialByAdmin updates the social networks of a user.
	UpdateSocialByAdmin(ctx context.Context, userID, facebook, twitter, instagram string) error
	// UpdateSettingsByAdmin updates the locale and the timezone of a user.
	UpdateSettingsByAdmin(ctx context.Context, userID, locale, timezone string) error
	// UpdateNotificationsByAdmin updates the notifications of a user.
	UpdateNotificationsByAdmin(ctx context.Context, userID string, notifications []string) error

	// GetTokensForAdmin returns the list of the auth tokens of a user to display in the admin panel.
	GetTokensForAdmin(ctx context.Context, userID string) (any, error)
	// DeleteTokenByIDByAdmin deletes an auth token of a user.
	DeleteTokenByIDByAdmin(ctx context.Context, tokenID string) error

	// SetAbilitiesToUser replaces the abilities of a user.
	SetAbilitiesToUser(ctx context.Context, userID string, abilities []string) error
	// AddAbilitiesToUser adds some abilities to a user.
	AddAbilitiesToUser(ctx context.Context, userID string, abilities []string) error
	// RemoveAbilitiesFromUser removes some abilities from a user.
	RemoveAbilitiesFromUser(ctx context.Context, userID string, abilities []string) error
}

// Inject returns a middleware saving the storage in the Gin context, see StorageOf.
func Inject(storage Storage) gin.HandlerFunc {
	return func(c *gin.Context) {
		SetStorage(c, storage)
		c.Next()
	}
}

// SetStorage saves the storage in the Gin context, like Inject does. It is meant for the tests.
func SetStorage(c *gin.Context, storage Storage) {
	c.Set(storageKey, storage)
}

// StorageOf returns the storage saved in the Gin context by Inject, or nil if the application provides none: the
// feature relying on it is disabled.
func StorageOf(c *gin.Context) Storage {
	storage, _ := c.Get(storageKey)
	s, _ := storage.(Storage)

	return s
}
