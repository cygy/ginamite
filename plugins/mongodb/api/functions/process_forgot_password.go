package functions

import (
	"errors"

	commonErrors "github.com/cygy/ginamite/common/errors"
	"github.com/cygy/ginamite/common/mongo/session"
	"github.com/cygy/ginamite/plugins/mongodb/model"

	"github.com/gin-gonic/gin"
	"github.com/globalsign/mgo/bson"
)

// NewForgotPasswordProcessForUser : returns the unique ID of a new forgot passsword process.
func NewForgotPasswordProcessForUser(c *gin.Context, userID, processPrivateKey string) (string, error) {
	mongoSession := session.Get(c)

	// Delete the existing forgot password process.
	if err := model.DeleteForgotPasswordProcessByUserID(userID, mongoSession); err != nil && !commonErrors.IsNotFound(err) {
		return "", err
	}

	// Create a new forgot password process.
	fpp := model.NewForgotPasswordProcess(userID)
	fpp.Code = processPrivateKey
	fppID, err := fpp.Save(mongoSession)
	if err != nil {
		return "", err
	}

	return fppID, nil
}

// GetForgotPasswordProcessDetailsByID : returns the user ID and the private key a forgot passsword process.
func GetForgotPasswordProcessDetailsByID(c *gin.Context, processID string) (userID, processPrivateKey string, err error) {
	if !bson.IsObjectIdHex(processID) {
		return "", "", errors.New("wrong forgot password process ID")
	}

	mongoSession := session.Get(c)

	fpp := &model.ForgotPasswordProcess{}
	if err := fpp.GetByID(processID, mongoSession); err != nil || !fpp.IsSaved() {
		return "", "", errors.New("unable to find the forgot password process")
	}

	return fpp.UserID.Hex(), fpp.Code, nil
}

// WrongAccessToForgotPasswordProcess : actions to execute when a wrong access to a forgot password process occurs.
func WrongAccessToForgotPasswordProcess(c *gin.Context, processID string) {
	mongoSession := session.Get(c)

	fpp := &model.ForgotPasswordProcess{}
	if err := fpp.GetByID(processID, mongoSession); err != nil || !fpp.IsSaved() {
		return
	}

	if fpp.Tries >= 2 {
		fpp.Delete(mongoSession)
	} else {
		fpp.Increment(mongoSession)
	}
}

// DeleteForgotPasswordProcess : deletes a forgot password process.
func DeleteForgotPasswordProcess(c *gin.Context, processID string) error {
	mongoSession := session.Get(c)

	process := &model.ForgotPasswordProcess{
		ID: bson.ObjectIdHex(processID),
	}

	return process.Delete(mongoSession)
}
