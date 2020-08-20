package model

import (
	"time"

	"github.com/cygy/ginamite/api/validator"
	"github.com/cygy/ginamite/common/authentication"

	"github.com/globalsign/mgo"
)

// SaveFacebookInfos : saves the facebook access token to the document.
func (user *User) SaveFacebookInfos(token, userID, firstName, lastName, email string, createdAt, expiresAt time.Time, db *mgo.Database) error {
	return user.saveThirdPartyInfos(authentication.MethodFacebook, token, userID, firstName, lastName, email, createdAt, expiresAt, db)
}

// SaveGoogleInfos : saves the google access token to the document.
func (user *User) SaveGoogleInfos(token, userID, firstName, lastName, email string, createdAt, expiresAt time.Time, db *mgo.Database) error {
	return user.saveThirdPartyInfos(authentication.MethodGoogle, token, userID, firstName, lastName, email, createdAt, expiresAt, db)
}

// FromSocialNetworkOAuthInfos : fills a UserOAuthInfos struct from a TokenInfos struct
func (oauthInfos *UserOAuthInfos) FromSocialNetworkOAuthInfos(tokenInfos validator.TokenInfos) {
	oauthInfos.UserID = tokenInfos.UserID
	oauthInfos.Token = tokenInfos.Token
	oauthInfos.FirstName = tokenInfos.FirstName
	oauthInfos.LastName = tokenInfos.LastName
	oauthInfos.Email = tokenInfos.Email
	oauthInfos.CreatedAt = tokenInfos.CreatedAt
	oauthInfos.ExpiresAt = tokenInfos.ExpiresAt
}
