package functions

import (
	"fmt"
	"time"

	"github.com/cygy/ginamite/common/authentication"
	"github.com/cygy/ginamite/common/errors"
	"github.com/cygy/ginamite/common/mongo/session"
	"github.com/cygy/ginamite/plugins/mongodb/model"

	"github.com/gin-gonic/gin"
	"github.com/globalsign/mgo/bson"
)

// BuildTokenFromID : returns the content of an auth token.
func BuildTokenFromID(tokenID string) (*authentication.Token, error) {
	session, db := session.Copy()
	defer session.Close()

	// Get the token.
	authToken := &model.AuthToken{}
	if err := authToken.GetByID(tokenID, db); err != nil {
		return nil, err
	}

	// Get the user.
	userID := authToken.UserID.Hex()
	user := &model.User{}
	if err := user.GetByID(userID, db); err != nil {
		return nil, err
	}

	return &authentication.Token{
		Source:         fmt.Sprintf("%s (%s)", authToken.Source, authToken.Method),
		UserID:         user.ID.Hex(),
		ExpirationDate: authToken.ExpiresAt,
		CreationDate:   authToken.CreatedAt,
		Other: map[string]interface{}{
			"nickname": user.Username,
			"locale":   user.Settings.Locale,
			"timezone": user.Settings.Timezone,
			"img1":     user.PublicInfos.Image.Thumb.Size1x.URL,
			"img2":     user.PublicInfos.Image.Thumb.Size2x.URL,
			"unread":   user.NotificationsStats.Unread,
		},
	}, nil
}

// ExtendTokenExpirationDateFromID : extends the expiration date of an auth token.
func ExtendTokenExpirationDateFromID(tokenID string, ttl time.Duration) error {
	session, db := session.Copy()
	defer session.Close()

	// Get the token.
	authToken := &model.AuthToken{}
	if err := authToken.GetByID(tokenID, db); err != nil {
		return err
	}

	// Extends the validity of the token.
	if authToken.ExpiresAt.Sub(time.Now()) < (ttl*time.Second)/2 {
		authToken.ExpiresAt = time.Now().Add(ttl * time.Second)
		if err := authToken.UpdateExpirationDate(authToken.ExpiresAt, db); err != nil {
			return err
		}
	}

	user := model.User{
		ID: authToken.UserID,
	}
	user.UpdateLastLogin(db)

	return nil
}

// SaveAuthenticationToken : saves an auth token.
func SaveAuthenticationToken(c *gin.Context, details authentication.Details) (tokenID string, knownIPAddress bool, err error) {
	mongoSession := session.Get(c)

	// Get the IP location info.
	location := &model.IPLocation{}
	if err := location.GetByIPAddress(details.IPAddress, mongoSession); err != nil {
		location = nil
	} else {
		location.ID = "" // Do not save the _id of the document.
	}

	device := model.Device{}
	device.FromDevice(details.Device)

	userID := bson.ObjectIdHex(details.UserID)

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

	authTokenID, err := authToken.Save(mongoSession)
	if err != nil {
		return "", true, err
	}

	user := model.User{
		ID: userID,
	}
	user.UpdateLastLogin(mongoSession)

	return authTokenID, (location != nil), nil
}

// GetTokenExpirationDate : returns the expiration date of an auth token.
func GetTokenExpirationDate(c *gin.Context, tokenID string) (time.Time, error) {
	mongoSession := session.Get(c)
	token := &model.AuthToken{}
	err := token.GetByID(tokenID, mongoSession)
	return token.ExpiresAt, err
}

// DeleteTokenByID : delete an auth token by its ID.
func DeleteTokenByID(c *gin.Context, tokenID string) error {
	mongoSession := session.Get(c)
	return model.DeleteAuthTokenByID(tokenID, mongoSession)
}

// GetTokenDetailsByID : returns the details of an auth token.
func GetTokenDetailsByID(c *gin.Context, tokenID string) (authentication.Token, error) {
	mongoSession := session.Get(c)
	token := &model.AuthToken{}
	err := token.GetByID(tokenID, mongoSession)
	return token.ToToken(), err
}

// GetOwnedTokens : returns the auth token for an user.
func GetOwnedTokens(c *gin.Context, userID string) (interface{}, error) {
	mongoSession := session.Get(c)

	tokens, err := model.GetAuthTokens(userID, mongoSession)
	if err != nil && !errors.IsNotFound(err) {
		return nil, err
	}

	ownTokens := make([]*model.OwnAuthToken, len(tokens))
	for i, token := range tokens {
		ownTokens[i] = model.NewOwnAuthToken(&token)
	}

	return ownTokens, nil
}

// GetOwnedTokenByID : returns an auth token by its ID for an user.
func GetOwnedTokenByID(c *gin.Context, tokenID string) (interface{}, string, error) {
	mongoSession := session.Get(c)

	token := &model.AuthToken{}
	if err := token.GetByID(tokenID, mongoSession); err != nil {
		return nil, "", err
	}

	return model.NewOwnAuthToken(token), token.UserID.Hex(), nil
}

// UpdateTokenByID : updates the details of a token.
func UpdateTokenByID(c *gin.Context, tokenID, name string, enableNotifications bool, notifications []string) error {
	mongoSession := session.Get(c)

	token := &model.AuthToken{
		ID: bson.ObjectIdHex(tokenID),
	}

	return token.UpdateProperties(name, enableNotifications, notifications, mongoSession)
}
