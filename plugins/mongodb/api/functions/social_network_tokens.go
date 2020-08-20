package functions

import (
	"errors"

	"github.com/cygy/ginamite/api/validator"
	"github.com/cygy/ginamite/common/mongo/session"
	"github.com/cygy/ginamite/plugins/mongodb/model"

	"github.com/gin-gonic/gin"
	"github.com/globalsign/mgo/bson"
)

// SaveFacebookUserTokenDetailsToUser : saves the user's facebook details of an user.
func SaveFacebookUserTokenDetailsToUser(c *gin.Context, userID string, tokenInfos validator.TokenInfos) error {
	// Check the userID.
	if !bson.IsObjectIdHex(userID) {
		return errors.New("wrong object ID")
	}

	mongoSession := session.Get(c)

	user := &model.User{}
	if err := user.GetByID(userID, mongoSession); err != nil {
		return err
	}

	return user.SaveFacebookInfos(tokenInfos.Token, tokenInfos.UserID, tokenInfos.FirstName, tokenInfos.LastName, tokenInfos.Email, tokenInfos.CreatedAt, tokenInfos.ExpiresAt, mongoSession)
}

// SaveGoogleUserTokenDetailsToUser : saves the user's google details of an user.
func SaveGoogleUserTokenDetailsToUser(c *gin.Context, userID string, tokenInfos validator.TokenInfos) error {
	// Check the userID.
	if !bson.IsObjectIdHex(userID) {
		return errors.New("wrong object ID")
	}

	mongoSession := session.Get(c)

	user := &model.User{}
	if err := user.GetByID(userID, mongoSession); err != nil {
		return err
	}

	return user.SaveGoogleInfos(tokenInfos.Token, tokenInfos.UserID, tokenInfos.FirstName, tokenInfos.LastName, tokenInfos.Email, tokenInfos.CreatedAt, tokenInfos.ExpiresAt, mongoSession)
}
