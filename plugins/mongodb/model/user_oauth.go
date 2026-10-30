package model

import (
	"context"
	"time"

	"github.com/cygy/ginamite/api/validator"
	"github.com/cygy/ginamite/common/authentication"
	"github.com/cygy/ginamite/common/mongo/store"
)

// SaveFacebookInfos saves the facebook access token to the document.
func (user *User) SaveFacebookInfos(ctx context.Context, token, userID, firstName, lastName, email string, createdAt, expiresAt time.Time, db store.Database) error {
	return user.saveThirdPartyInfos(ctx, authentication.MethodFacebook, token, userID, firstName, lastName, email, createdAt, expiresAt, db)
}

// SaveGoogleInfos saves the google access token to the document.
func (user *User) SaveGoogleInfos(ctx context.Context, token, userID, firstName, lastName, email string, createdAt, expiresAt time.Time, db store.Database) error {
	return user.saveThirdPartyInfos(ctx, authentication.MethodGoogle, token, userID, firstName, lastName, email, createdAt, expiresAt, db)
}

// FromSocialNetworkOAuthInfos fills the UserOAuthInfos from the information of a social network token.
func (oauthInfos *UserOAuthInfos) FromSocialNetworkOAuthInfos(tokenInfos validator.TokenInfos) {
	oauthInfos.UserID = tokenInfos.UserID
	oauthInfos.Token = tokenInfos.Token
	oauthInfos.FirstName = tokenInfos.FirstName
	oauthInfos.LastName = tokenInfos.LastName
	oauthInfos.Email = tokenInfos.Email
	oauthInfos.CreatedAt = tokenInfos.CreatedAt
	oauthInfos.ExpiresAt = tokenInfos.ExpiresAt
}
