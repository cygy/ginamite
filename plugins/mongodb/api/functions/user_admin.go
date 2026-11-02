package functions

import (
	"context"

	"github.com/cygy/ginamite/common/errors"
	"github.com/cygy/ginamite/plugins/mongodb/model"
	"github.com/cygy/ginamite/plugins/mongodb/mongo/store"
)

// Users stores the users managed by the administrators in a database. It is a user.Storage. Create it with NewUsers.
type Users struct {
	database func() store.Database
	settings func() model.Settings
}

// NewUsers returns the storage of the users working on the database returned by the function, which is called each
// time the storage needs the database. The abilities which can be given to the users are the ones of the settings.
func NewUsers(database func() store.Database, settings func() model.Settings) *Users {
	return &Users{database: database, settings: settings}
}

// GetEnabledUsersForAdmin returns the enabled users for the admin panel.
func (s *Users) GetEnabledUsersForAdmin(ctx context.Context) (any, error) {
	users, err := model.GetEnabledUsersForAdmin(ctx, s.database())

	return users, err
}

// GetDisabledUsersForAdmin returns the disabled users for the admin panel.
func (s *Users) GetDisabledUsersForAdmin(ctx context.Context) (any, error) {
	users, err := model.GetDisabledUsersForAdmin(ctx, s.database())

	return users, err
}

// UpdatePasswordByAdmin updates the password of a user.
func (s *Users) UpdatePasswordByAdmin(ctx context.Context, userID, password string) error {
	user, err := model.UserWithID(userID)
	if err != nil {
		return err
	}

	return user.UpdatePassword(ctx, password, false, s.database())
}

// UpdateInformationsByAdmin updates the informations of a user.
func (s *Users) UpdateInformationsByAdmin(ctx context.Context, userID, firstName, lastName, emailAddress string, validEmailAddress bool) error {
	user, err := model.UserWithID(userID)
	if err != nil {
		return err
	}

	return user.UpdateAllInformations(ctx, firstName, lastName, emailAddress, validEmailAddress, s.database())
}

// UpdateDescriptionByAdmin updates the description of a user.
func (s *Users) UpdateDescriptionByAdmin(ctx context.Context, userID, locale, description string) error {
	user, err := model.UserWithID(userID)
	if err != nil {
		return err
	}

	return user.UpdatePublicProfile(ctx, locale, description, s.database())
}

// UpdateSocialByAdmin updates the social networks of a user.
func (s *Users) UpdateSocialByAdmin(ctx context.Context, userID, facebook, twitter, instagram string) error {
	user, err := model.UserWithID(userID)
	if err != nil {
		return err
	}

	return user.UpdateSocialNetworks(ctx, facebook, twitter, instagram, s.database())
}

// UpdateSettingsByAdmin updates the locale and the timezone of a user.
func (s *Users) UpdateSettingsByAdmin(ctx context.Context, userID, locale, timezone string) error {
	user, err := model.UserWithID(userID)
	if err != nil {
		return err
	}

	return user.UpdateSettings(ctx, locale, timezone, s.database())
}

// UpdateNotificationsByAdmin updates the notifications of a user.
func (s *Users) UpdateNotificationsByAdmin(ctx context.Context, userID string, notifications []string) error {
	user, err := model.UserWithID(userID)
	if err != nil {
		return err
	}

	return user.UpdateNotifications(ctx, notifications, s.database())
}

// GetTokensForAdmin returns the auth tokens of the user, as model.AdminAuthToken, for the admin panel.
func (s *Users) GetTokensForAdmin(ctx context.Context, userID string) (any, error) {
	tokens, err := model.GetAuthTokens(ctx, userID, s.database())
	if err != nil && !errors.IsNotFound(err) {
		return nil, err
	}

	adminTokens := make([]*model.AdminAuthToken, len(tokens))
	for i, token := range tokens {
		adminTokens[i] = model.NewAdminAuthToken(&token)
	}

	return adminTokens, nil
}

// DeleteTokenByIDByAdmin deletes an auth token of a user.
func (s *Users) DeleteTokenByIDByAdmin(ctx context.Context, tokenID string) error {
	return model.DeleteAuthTokenByID(ctx, tokenID, s.database())
}

// SetAbilitiesToUser replaces the abilities of a user.
func (s *Users) SetAbilitiesToUser(ctx context.Context, userID string, abilities []string) error {
	user, err := model.UserWithID(userID)
	if err != nil {
		return err
	}

	return user.SetAbilities(ctx, abilities, s.settings().AllUserAbilities, s.database())
}

// AddAbilitiesToUser adds some abilities to a user.
func (s *Users) AddAbilitiesToUser(ctx context.Context, userID string, abilities []string) error {
	user, err := model.UserWithID(userID)
	if err != nil {
		return err
	}

	return user.AddAbilities(ctx, abilities, s.settings().AllUserAbilities, s.database())
}

// RemoveAbilitiesFromUser removes some abilities from a user.
func (s *Users) RemoveAbilitiesFromUser(ctx context.Context, userID string, abilities []string) error {
	user, err := model.UserWithID(userID)
	if err != nil {
		return err
	}

	return user.RemoveAbilities(ctx, abilities, s.database())
}
