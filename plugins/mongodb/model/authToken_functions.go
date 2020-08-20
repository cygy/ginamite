package model

import (
	"time"

	"github.com/cygy/ginamite/common/mongo/document"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

// GetAuthTokens : returns all the auth tokens of a user.
func GetAuthTokens(userID string, db *mgo.Database) ([]AuthToken, error) {
	tokens := []AuthToken{}
	err := document.GetDocumentsBySelectorAndSort(&tokens, AuthTokenCollection, bson.M{"userId": bson.ObjectIdHex(userID)}, []string{"userId", "-createdAt"}, db)

	return tokens, err
}

// UserFromAuthToken : returns a user with the values of a auth token.
func UserFromAuthToken(token *jwt.Token) User {
	claims := token.Claims.(jwt.MapClaims)

	var userID bson.ObjectId
	var nickname, locale, timezone string
	if value, ok := claims["sub"]; ok {
		userID = bson.ObjectIdHex(value.(string))
	}
	if value, ok := claims["nickname"]; ok {
		nickname = value.(string)
	}
	if value, ok := claims["locale"]; ok {
		locale = value.(string)
	}
	if value, ok := claims["timezone"]; ok {
		timezone = value.(string)
	}

	return User{
		ID:       userID,
		Username: nickname,
		Settings: UserSettings{
			Locale:   locale,
			Timezone: timezone,
		},
	}
}

// DeleteAuthTokenByID : deletes the document with the ID.
func DeleteAuthTokenByID(tokenID string, db *mgo.Database) error {
	return document.DeleteDocument(bson.ObjectIdHex(tokenID), AuthTokenCollection, db)
}

// DeleteAuthTokensByUser : deletes the documents from the user.
func DeleteAuthTokensByUser(userID string, db *mgo.Database) (int, error) {
	return document.DeleteDocumentsBySelector(AuthTokenCollection, bson.M{"userId": bson.ObjectIdHex(userID)}, db)
}

// DeleteExpiredAuthTokens : deletes the expired auth tokens.
func DeleteExpiredAuthTokens(db *mgo.Database) (int, error) {
	return document.DeleteDocumentsBySelector(AuthTokenCollection, bson.M{"expiresAt": bson.M{"$lt": time.Now()}}, db)
}

// IsValidAuthToken : checks that the auth token is valid (non-expired and non-revoked).
func IsValidAuthToken(tokenID string, db *mgo.Database) bool {
	count, _ := document.CountBySelector(AuthTokenCollection, bson.D{{Name: "_id", Value: bson.ObjectIdHex(tokenID)}, {Name: "expiresAt", Value: bson.M{"$gt": time.Now()}}}, db)
	return count == 1
}
