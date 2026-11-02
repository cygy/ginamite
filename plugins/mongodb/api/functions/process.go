package functions

import (
	"context"
	"crypto/subtle"

	"github.com/cygy/ginamite/common/errors"
	"github.com/cygy/ginamite/common/log"
	"github.com/cygy/ginamite/plugins/mongodb/model"
	"github.com/cygy/ginamite/plugins/mongodb/mongo/bson"
	"github.com/cygy/ginamite/plugins/mongodb/mongo/session"
	"github.com/cygy/ginamite/plugins/mongodb/mongo/store"
	"github.com/gin-gonic/gin"
)

// CreateNewDeleteAccountProcess creates a process to delete the account of the user, replacing the existing one.
// It returns the unique ID of the process and its code.
func CreateNewDeleteAccountProcess(c *gin.Context, userID string) (string, string, error) {
	return createProcess(c, userID, model.DeleteDeleteAccountProcessByUserID, func() (*model.UpdatePropertyProcess, error) {
		return model.NewDeleteAccountProcess(userID)
	})
}

// CreateNewDisableAccountProcess creates a process to disable the account of the user, replacing the existing one.
// It returns the unique ID of the process and its code.
func CreateNewDisableAccountProcess(c *gin.Context, userID string) (string, string, error) {
	return createProcess(c, userID, model.DeleteDisableAccountProcessByUserID, func() (*model.UpdatePropertyProcess, error) {
		return model.NewDisableAccountProcess(userID)
	})
}

// CreateNewUpdatePasswordProcess creates a process to update the password of the user, replacing the existing one.
// It returns the unique ID of the process and its code.
func CreateNewUpdatePasswordProcess(c *gin.Context, userID, encryptedPassword string) (string, string, error) {
	return createProcess(c, userID, model.DeleteUpdatePasswordProcessByUserID, func() (*model.UpdatePropertyProcess, error) {
		return model.NewUpdatePasswordProcess(userID, encryptedPassword)
	})
}

// CreateNewUpdateEmailAddressProcess creates a process to update the email address of the user, replacing the
// existing one. It returns the unique ID of the process and its code.
func CreateNewUpdateEmailAddressProcess(c *gin.Context, userID, emailAddress string) (string, string, error) {
	return createProcess(c, userID, model.DeleteUpdateEmailAddressProcessByUserID, func() (*model.UpdatePropertyProcess, error) {
		return model.NewUpdateEmailAddressProcess(userID, emailAddress)
	})
}

// CreateNewVerifyEmailAddressProcess creates a process to verify the email address of the user, replacing the
// existing one. It returns the unique ID of the process and its code.
func CreateNewVerifyEmailAddressProcess(c *gin.Context, userID string) (string, string, error) {
	return createProcess(c, userID, model.DeleteUpdateValidEmailAddressProcessByUserID, func() (*model.UpdatePropertyProcess, error) {
		return model.NewUpdateValidEmailAddressProcess(userID)
	})
}

// createProcess deletes the existing process of the user, saves the new one and returns its ID and its code.
func createProcess(c *gin.Context, userID string, deleteExisting func(context.Context, string, store.Database) error, newProcess func() (*model.UpdatePropertyProcess, error)) (string, string, error) {
	mongoSession := session.Get(c)

	if err := deleteExisting(c, userID, mongoSession); err != nil && !errors.IsNotFound(err) {
		return "", "", err
	}

	process, err := newProcess()
	if err != nil {
		return "", "", err
	}

	processID, err := process.Save(c, mongoSession)
	if err != nil {
		return "", "", err
	}

	return processID, process.Code, nil
}

// VerifyProcess verifies the key of a process and returns the ID of its user and its stored value.
// A wrong key counts as a try: the process is deleted after three wrong tries. It returns a "not found" error if the
// process does not exist or if the key is wrong.
func VerifyProcess(c *gin.Context, processID, key string) (string, string, error) {
	if _, err := bson.ObjectIDFromHex(processID); err != nil {
		return "", "", err
	}

	mongoSession := session.Get(c)

	process := &model.UpdatePropertyProcess{}
	if err := process.GetByID(c, processID, mongoSession); err != nil || !process.IsSaved() {
		return "", "", errors.NotFound()
	}

	// Do not validate or cancel if the key is not verified.
	if subtle.ConstantTimeCompare([]byte(process.Code), []byte(key)) != 1 {
		if process.Tries >= 2 {
			if err := process.Delete(c, mongoSession); err != nil {
				log.WithField("error", err).Error("unable to delete")
			}
		} else {
			if err := process.Increment(c, mongoSession); err != nil {
				log.WithField("error", err).Error("unable to increment")
			}
		}
		return "", "", errors.NotFound()
	}

	return process.UserID.Hex(), process.Value, nil
}

// DeleteProcessByID deletes a process by its unique ID.
func DeleteProcessByID(c *gin.Context, processID string) error {
	objectID, err := bson.ObjectIDFromHex(processID)
	if err != nil {
		return err
	}

	mongoSession := session.Get(c)

	process := &model.UpdatePropertyProcess{
		ID: objectID,
	}

	return process.Delete(c, mongoSession)
}
