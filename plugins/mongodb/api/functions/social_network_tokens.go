package functions

import (
	"context"

	"github.com/cygy/ginamite/api/validator"
	"github.com/cygy/ginamite/plugins/mongodb/model"
	"github.com/cygy/ginamite/plugins/mongodb/mongo/bson"
)

// SaveFacebookUserTokenDetailsToUser saves the user's facebook details of a user.
func (a *Accounts) SaveFacebookUserTokenDetailsToUser(ctx context.Context, userID string, tokenInfos validator.TokenInfos) error {
	// Check the userID.
	if _, err := bson.ObjectIDFromHex(userID); err != nil {
		return err
	}

	mongoSession := a.database()

	user := &model.User{}
	if err := user.GetByID(ctx, userID, mongoSession); err != nil {
		return err
	}

	return user.SaveFacebookInfos(ctx, tokenInfos.Token, tokenInfos.UserID, tokenInfos.FirstName, tokenInfos.LastName, tokenInfos.Email, tokenInfos.CreatedAt, tokenInfos.ExpiresAt, mongoSession)
}

// SaveGoogleUserTokenDetailsToUser saves the user's google details of a user.
func (a *Accounts) SaveGoogleUserTokenDetailsToUser(ctx context.Context, userID string, tokenInfos validator.TokenInfos) error {
	// Check the userID.
	if _, err := bson.ObjectIDFromHex(userID); err != nil {
		return err
	}

	mongoSession := a.database()

	user := &model.User{}
	if err := user.GetByID(ctx, userID, mongoSession); err != nil {
		return err
	}

	return user.SaveGoogleInfos(ctx, tokenInfos.Token, tokenInfos.UserID, tokenInfos.FirstName, tokenInfos.LastName, tokenInfos.Email, tokenInfos.CreatedAt, tokenInfos.ExpiresAt, mongoSession)
}
