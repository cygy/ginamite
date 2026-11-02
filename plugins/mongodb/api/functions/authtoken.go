package functions

import (
	"context"
	"fmt"
	"time"

	"github.com/cygy/ginamite/common/authentication"
	"github.com/cygy/ginamite/common/errors"
	"github.com/cygy/ginamite/common/log"
	"github.com/cygy/ginamite/plugins/mongodb/model"
	"github.com/cygy/ginamite/plugins/mongodb/mongo/bson"
	"github.com/cygy/ginamite/plugins/mongodb/mongo/store"
)

// Tokens builds, extends and manages the auth tokens stored in a database. It is an account.TokensStorage. Create it
// with NewTokens.
type Tokens struct {
	database func() store.Database
}

// NewTokens returns the functions of the auth tokens working on the database returned by the function, which is called
// each time a function needs the database.
func NewTokens(database func() store.Database) *Tokens {
	return &Tokens{database: database}
}

// BuildTokenFromID builds the authentication.Token of an auth token, with the details of its user.
func (t *Tokens) BuildTokenFromID(ctx context.Context, tokenID string) (*authentication.Token, error) {
	db := t.database()

	// Get the token.
	authToken := &model.AuthToken{}
	if err := authToken.GetByID(ctx, tokenID, db); err != nil {
		return nil, err
	}

	// Get the user.
	userID := authToken.UserID.Hex()
	user := &model.User{}
	if err := user.GetByID(ctx, userID, db); err != nil {
		return nil, err
	}

	return &authentication.Token{
		Source:         fmt.Sprintf("%s (%s)", authToken.Source, authToken.Method),
		UserID:         user.ID.Hex(),
		ExpirationDate: authToken.ExpiresAt,
		CreationDate:   authToken.CreatedAt,
		Other: map[string]any{
			"nickname": user.Username,
			"locale":   user.Settings.Locale,
			"timezone": user.Settings.Timezone,
			"img1":     user.PublicInfos.Image.Thumb.Size1x.URL,
			"img2":     user.PublicInfos.Image.Thumb.Size2x.URL,
			"unread":   user.NotificationsStats.Unread,
		},
	}, nil
}

// ExtendTokenExpirationDateFromID extends the validity of an auth token by ttl seconds, if less than half of the
// ttl is left, and updates the last login date of its user.
func (t *Tokens) ExtendTokenExpirationDateFromID(ctx context.Context, tokenID string, ttl time.Duration) error {
	db := t.database()

	// Get the token.
	authToken := &model.AuthToken{}
	if err := authToken.GetByID(ctx, tokenID, db); err != nil {
		return err
	}

	// Extends the validity of the token.
	if time.Until(authToken.ExpiresAt) < (ttl*time.Second)/2 {
		authToken.ExpiresAt = time.Now().Add(ttl * time.Second)
		if err := authToken.UpdateExpirationDate(ctx, authToken.ExpiresAt, db); err != nil {
			return err
		}
	}

	user := model.User{
		ID: authToken.UserID,
	}
	if err := user.UpdateLastLogin(ctx, db); err != nil {
		log.WithField("error", err).Error("unable to update last login")
	}

	return nil
}

// SaveAuthenticationToken saves an auth token.
func (t *Tokens) SaveAuthenticationToken(ctx context.Context, details authentication.Details) (tokenID string, knownIPAddress bool, err error) {
	userID, err := bson.ObjectIDFromHex(details.UserID)
	if err != nil {
		return "", false, err
	}

	db := t.database()

	// Get the IP location info.
	location := &model.IPLocation{}
	if err := location.GetByIPAddress(ctx, details.IPAddress, db); err != nil {
		location = nil
	} else {
		location.ID = bson.NewObjectID() // Do not save the _id of the document.
	}

	device := model.Device{}
	device.FromDevice(details.Device)

	authToken := model.AuthToken{
		UserID:    userID,
		Source:    details.Source,
		Method:    details.Method,
		Device:    device,
		Location:  location,
		Key:       details.Key,
		CreatedAt: details.CreationDate,
		ExpiresAt: details.ExpirationDate,
	}

	authTokenID, err := authToken.Save(ctx, db)
	if err != nil {
		return "", true, err
	}

	user := model.User{
		ID: userID,
	}
	if err := user.UpdateLastLogin(ctx, db); err != nil {
		log.WithField("error", err).Error("unable to update last login")
	}

	return authTokenID, (location != nil), nil
}

// GetTokenExpirationDate returns the expiration date of an auth token.
func (t *Tokens) GetTokenExpirationDate(ctx context.Context, tokenID string) (time.Time, error) {
	token := &model.AuthToken{}
	err := token.GetByID(ctx, tokenID, t.database())

	return token.ExpiresAt, err
}

// DeleteTokenByID deletes an auth token by its ID.
func (t *Tokens) DeleteTokenByID(ctx context.Context, tokenID string) error {
	return model.DeleteAuthTokenByID(ctx, tokenID, t.database())
}

// GetTokenDetailsByID returns the details of an auth token.
func (t *Tokens) GetTokenDetailsByID(ctx context.Context, tokenID string) (authentication.Token, error) {
	token := &model.AuthToken{}
	err := token.GetByID(ctx, tokenID, t.database())

	return token.ToToken(), err
}

// GetOwnedTokens returns the auth tokens of the user, as model.OwnAuthToken.
func (t *Tokens) GetOwnedTokens(ctx context.Context, userID string) (any, error) {
	tokens, err := model.GetAuthTokens(ctx, userID, t.database())
	if err != nil && !errors.IsNotFound(err) {
		return nil, err
	}

	ownTokens := make([]*model.OwnAuthToken, len(tokens))
	for i, token := range tokens {
		ownTokens[i] = model.NewOwnAuthToken(&token)
	}

	return ownTokens, nil
}

// GetOwnedTokenByID returns an auth token, as a model.OwnAuthToken, and the ID of its user.
func (t *Tokens) GetOwnedTokenByID(ctx context.Context, tokenID string) (any, string, error) {
	token := &model.AuthToken{}
	if err := token.GetByID(ctx, tokenID, t.database()); err != nil {
		return nil, "", err
	}

	return model.NewOwnAuthToken(token), token.UserID.Hex(), nil
}

// UpdateTokenByID updates the details of a token.
func (t *Tokens) UpdateTokenByID(ctx context.Context, tokenID, name string, enableNotifications bool, notifications []string) error {
	objectID, err := bson.ObjectIDFromHex(tokenID)
	if err != nil {
		return err
	}

	token := &model.AuthToken{
		ID: objectID,
	}

	return token.UpdateProperties(ctx, name, enableNotifications, notifications, t.database())
}
