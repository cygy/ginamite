package functions

import (
	"errors"

	commonErrors "github.com/cygy/ginamite/common/errors"
	"github.com/cygy/ginamite/common/log"
	"github.com/cygy/ginamite/common/mongo/bson"
	"github.com/cygy/ginamite/common/mongo/session"
	"github.com/cygy/ginamite/plugins/mongodb/model"
	"github.com/gin-gonic/gin"
)

// NewForgotPasswordProcessForUser : returns the unique ID of a new forgot passsword process.
func NewForgotPasswordProcessForUser(c *gin.Context, userID, processPrivateKey string) (string, error) {
	mongoSession := session.Get(c)

	// Delete the existing forgot password process.
	if err := model.DeleteForgotPasswordProcessByUserID(c, userID, mongoSession); err != nil && !commonErrors.IsNotFound(err) {
		return "", err
	}

	// Create a new forgot password process.
	fpp, err := model.NewForgotPasswordProcess(userID)
	if err != nil {
		return "", err
	}
	fpp.Code = processPrivateKey
	fppID, err := fpp.Save(c, mongoSession)
	if err != nil {
		return "", err
	}

	return fppID, nil
}

// GetForgotPasswordProcessDetailsByID : returns the user ID and the private key a forgot passsword process.
func GetForgotPasswordProcessDetailsByID(c *gin.Context, processID string) (userID, processPrivateKey string, err error) {
	if _, err := bson.ObjectIDFromHex(processID); err != nil {
		return "", "", errors.New("wrong forgot password process ID")
	}

	mongoSession := session.Get(c)

	fpp := &model.ForgotPasswordProcess{}
	if err := fpp.GetByID(c, processID, mongoSession); err != nil || !fpp.IsSaved() {
		return "", "", errors.New("unable to find the forgot password process")
	}

	return fpp.UserID.Hex(), fpp.Code, nil
}

// WrongAccessToForgotPasswordProcess : actions to execute when a wrong access to a forgot password process occurs.
func WrongAccessToForgotPasswordProcess(c *gin.Context, processID string) {
	mongoSession := session.Get(c)

	fpp := &model.ForgotPasswordProcess{}
	if err := fpp.GetByID(c, processID, mongoSession); err != nil || !fpp.IsSaved() {
		return
	}

	if fpp.Tries >= 2 {
		if err := fpp.Delete(c, mongoSession); err != nil {
			log.WithField("error", err).Error("unable to delete")
		}
	} else {
		if err := fpp.Increment(c, mongoSession); err != nil {
			log.WithField("error", err).Error("unable to increment")
		}
	}
}

// DeleteForgotPasswordProcess : deletes a forgot password process.
func DeleteForgotPasswordProcess(c *gin.Context, processID string) error {
	objectID, err := bson.ObjectIDFromHex(processID)
	if err != nil {
		return err
	}

	mongoSession := session.Get(c)

	process := &model.ForgotPasswordProcess{
		ID: objectID,
	}

	return process.Delete(c, mongoSession)
}
