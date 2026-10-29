package functions

import (
	"github.com/cygy/ginamite/api/validator"
	"github.com/cygy/ginamite/common/mongo/bson"
	"github.com/cygy/ginamite/common/mongo/session"
	"github.com/cygy/ginamite/plugins/mongodb/model"
	"github.com/gin-gonic/gin"
)

// SaveFacebookUserTokenDetailsToUser : saves the user's facebook details of an user.
func SaveFacebookUserTokenDetailsToUser(c *gin.Context, userID string, tokenInfos validator.TokenInfos) error {
	// Check the userID.
	if _, err := bson.ObjectIDFromHex(userID); err != nil {
		return err
	}

	mongoSession := session.Get(c)

	user := &model.User{}
	if err := user.GetByID(c, userID, mongoSession); err != nil {
		return err
	}

	return user.SaveFacebookInfos(c, tokenInfos.Token, tokenInfos.UserID, tokenInfos.FirstName, tokenInfos.LastName, tokenInfos.Email, tokenInfos.CreatedAt, tokenInfos.ExpiresAt, mongoSession)
}

// SaveGoogleUserTokenDetailsToUser : saves the user's google details of an user.
func SaveGoogleUserTokenDetailsToUser(c *gin.Context, userID string, tokenInfos validator.TokenInfos) error {
	// Check the userID.
	if _, err := bson.ObjectIDFromHex(userID); err != nil {
		return err
	}

	mongoSession := session.Get(c)

	user := &model.User{}
	if err := user.GetByID(c, userID, mongoSession); err != nil {
		return err
	}

	return user.SaveGoogleInfos(c, tokenInfos.Token, tokenInfos.UserID, tokenInfos.FirstName, tokenInfos.LastName, tokenInfos.Email, tokenInfos.CreatedAt, tokenInfos.ExpiresAt, mongoSession)
}
