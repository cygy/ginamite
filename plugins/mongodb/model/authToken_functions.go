package model

import (
	"context"
	"time"

	"github.com/cygy/ginamite/common/mongo/bson"
	"github.com/cygy/ginamite/common/mongo/document"
	jwt "github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// GetAuthTokens : returns all the auth tokens of a user.
func GetAuthTokens(ctx context.Context, userID string, db *mongo.Database) ([]AuthToken, error) {
	tokens := []AuthToken{}

	objectID, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		return tokens, err
	}

	err = document.GetDocumentsBySelectorAndSort(ctx, &tokens, AuthTokenCollection, bson.M{"userId": objectID}, []string{"userId", "-createdAt"}, db)
	return tokens, err
}

// UserFromAuthToken : returns a user with the values of a auth token.
func UserFromAuthToken(token *jwt.Token) (*User, error) {
	claims := token.Claims.(jwt.MapClaims)

	var userID bson.ObjectID
	var nickname, locale, timezone string
	if value, ok := claims["sub"]; ok {
		objectID, err := bson.ObjectIDFromHex(value.(string))
		if err != nil {
			return nil, err
		}
		userID = objectID
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

	return &User{
		ID:       userID,
		Username: nickname,
		Settings: UserSettings{
			Locale:   locale,
			Timezone: timezone,
		},
	}, nil
}

// DeleteAuthTokenByID : deletes the document with the ID.
func DeleteAuthTokenByID(ctx context.Context, tokenID string, db *mongo.Database) error {
	objectID, err := bson.ObjectIDFromHex(tokenID)
	if err != nil {
		return err
	}
	return document.DeleteDocument(ctx, objectID, AuthTokenCollection, db)
}

// DeleteAuthTokensByUser : deletes the documents from the user.
func DeleteAuthTokensByUser(ctx context.Context, userID string, db *mongo.Database) (int64, error) {
	objectID, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		return 0, err
	}
	return document.DeleteDocumentsBySelector(ctx, AuthTokenCollection, bson.M{"userId": objectID}, db)
}

// DeleteExpiredAuthTokens : deletes the expired auth tokens.
func DeleteExpiredAuthTokens(ctx context.Context, db *mongo.Database) (int64, error) {
	return document.DeleteDocumentsBySelector(ctx, AuthTokenCollection, bson.M{"expiresAt": bson.M{"$lt": time.Now()}}, db)
}

// IsValidAuthToken : checks that the auth token is valid (non-expired and non-revoked).
func IsValidAuthToken(ctx context.Context, tokenID string, db *mongo.Database) (bool, error) {
	objectID, err := bson.ObjectIDFromHex(tokenID)
	if err != nil {
		return false, err
	}
	count, err := document.CountBySelector(ctx, AuthTokenCollection, bson.D{
		{Key: "_id", Value: objectID},
		{Key: "expiresAt", Value: bson.M{"$gt": time.Now()}},
	}, db)
	return count == 1, err
}
