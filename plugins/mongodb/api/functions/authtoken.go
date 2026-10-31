package functions

import (
	"context"
	"fmt"
	"time"

	"github.com/cygy/ginamite/common/authentication"
	"github.com/cygy/ginamite/common/errors"
	"github.com/cygy/ginamite/common/log"
	"github.com/cygy/ginamite/common/mongo/bson"
	"github.com/cygy/ginamite/common/mongo/session"
	"github.com/cygy/ginamite/common/mongo/store"
	"github.com/cygy/ginamite/plugins/mongodb/model"
	"github.com/gin-gonic/gin"
)

// Tokens builds and extends the auth tokens stored in a database. Create it with NewTokens.
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
func SaveAuthenticationToken(c *gin.Context, details authentication.Details) (tokenID string, knownIPAddress bool, err error) {
	userID, err := bson.ObjectIDFromHex(details.UserID)
	if err != nil {
		return "", false, err
	}

	mongoSession := session.Get(c)

	// Get the IP location info.
	location := &model.IPLocation{}
	if err := location.GetByIPAddress(c, details.IPAddress, mongoSession); err != nil {
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

	authTokenID, err := authToken.Save(c, mongoSession)
	if err != nil {
		return "", true, err
	}

	user := model.User{
		ID: userID,
	}
	if err := user.UpdateLastLogin(c, mongoSession); err != nil {
		log.WithField("error", err).Error("unable to update last login")
	}

	return authTokenID, (location != nil), nil
}

// GetTokenExpirationDate returns the expiration date of an auth token.
func GetTokenExpirationDate(c *gin.Context, tokenID string) (time.Time, error) {
	mongoSession := session.Get(c)
	token := &model.AuthToken{}
	err := token.GetByID(c, tokenID, mongoSession)
	return token.ExpiresAt, err
}

// DeleteTokenByID deletes an auth token by its ID.
func DeleteTokenByID(c *gin.Context, tokenID string) error {
	mongoSession := session.Get(c)
	return model.DeleteAuthTokenByID(c, tokenID, mongoSession)
}

// GetTokenDetailsByID returns the details of an auth token.
func GetTokenDetailsByID(c *gin.Context, tokenID string) (authentication.Token, error) {
	mongoSession := session.Get(c)
	token := &model.AuthToken{}
	err := token.GetByID(c, tokenID, mongoSession)
	return token.ToToken(), err
}

// GetOwnedTokens returns the auth tokens of the user, as model.OwnAuthToken.
func GetOwnedTokens(c *gin.Context, userID string) (any, error) {
	mongoSession := session.Get(c)

	tokens, err := model.GetAuthTokens(c, userID, mongoSession)
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
func GetOwnedTokenByID(c *gin.Context, tokenID string) (any, string, error) {
	mongoSession := session.Get(c)

	token := &model.AuthToken{}
	if err := token.GetByID(c, tokenID, mongoSession); err != nil {
		return nil, "", err
	}

	return model.NewOwnAuthToken(token), token.UserID.Hex(), nil
}

// UpdateTokenByID updates the details of a token.
func UpdateTokenByID(c *gin.Context, tokenID, name string, enableNotifications bool, notifications []string) error {
	objectID, err := bson.ObjectIDFromHex(tokenID)
	if err != nil {
		return err
	}
	mongoSession := session.Get(c)

	token := &model.AuthToken{
		ID: objectID,
	}

	return token.UpdateProperties(c, name, enableNotifications, notifications, mongoSession)
}
