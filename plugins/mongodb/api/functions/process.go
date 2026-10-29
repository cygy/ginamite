package functions

import (
	"github.com/cygy/ginamite/common/errors"
	"github.com/cygy/ginamite/common/mongo/session"
	"github.com/cygy/ginamite/plugins/mongodb/model"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// CreateNewDeleteAccountProcess : creates a process and returns its unique ID.
func CreateNewDeleteAccountProcess(c *gin.Context, userID string) (string, string, error) {
	mongoSession := session.Get(c)

	// Delete the existing process.
	if err := model.DeleteDeleteAccountProcessByUserID(c, userID, mongoSession); err != nil && !errors.IsNotFound(err) {
		return "", "", err
	}

	// Create a new process.
	process, err := model.NewDeleteAccountProcess(userID)
	if err != nil {
		return "", "", err
	}

	processID, err := process.Save(c, mongoSession)
	if err != nil {
		return "", "", err
	}

	return processID, process.Code, nil
}

// CreateNewDisableAccountProcess : creates a process and returns its unique ID.
func CreateNewDisableAccountProcess(c *gin.Context, userID string) (string, string, error) {
	mongoSession := session.Get(c)

	// Delete the existing process.
	if err := model.DeleteDisableAccountProcessByUserID(c, userID, mongoSession); err != nil && !errors.IsNotFound(err) {
		return "", "", err
	}

	// Create a new process.
	process, err := model.NewDisableAccountProcess(userID)
	if err != nil {
		return "", "", err
	}

	processID, err := process.Save(c, mongoSession)
	if err != nil {
		return "", "", err
	}

	return processID, process.Code, nil
}

// CreateNewUpdatePasswordProcess : creates a process and returns its unique ID.
func CreateNewUpdatePasswordProcess(c *gin.Context, userID, encryptedPassword string) (string, string, error) {
	mongoSession := session.Get(c)

	// Delete the existing process.
	if err := model.DeleteUpdatePasswordProcessByUserID(c, userID, mongoSession); err != nil && !errors.IsNotFound(err) {
		return "", "", err
	}

	// Create a new process.
	process, err := model.NewUpdatePasswordProcess(userID, encryptedPassword)
	if err != nil {
		return "", "", err
	}

	processID, err := process.Save(c, mongoSession)
	if err != nil {
		return "", "", err
	}

	return processID, process.Code, nil
}

// CreateNewUpdateEmailAddressProcess : creates a process and returns its unique ID.
func CreateNewUpdateEmailAddressProcess(c *gin.Context, userID, emailAddress string) (string, string, error) {
	mongoSession := session.Get(c)

	// Delete the existing process.
	if err := model.DeleteUpdateEmailAddressProcessByUserID(c, userID, mongoSession); err != nil && !errors.IsNotFound(err) {
		return "", "", err
	}

	// Create a new process.
	process, err := model.NewUpdateEmailAddressProcess(userID, emailAddress)
	if err != nil {
		return "", "", err
	}

	processID, err := process.Save(c, mongoSession)
	if err != nil {
		return "", "", err
	}

	return processID, process.Code, nil
}

// CreateNewVerifyEmailAddressProcess : creates a process and returns its unique ID.
func CreateNewVerifyEmailAddressProcess(c *gin.Context, userID string) (string, string, error) {
	mongoSession := session.Get(c)

	// Delete the existing process.
	if err := model.DeleteUpdateValidEmailAddressProcessByUserID(c, userID, mongoSession); err != nil && !errors.IsNotFound(err) {
		return "", "", err
	}

	// Create a new process.
	process, err := model.NewUpdateValidEmailAddressProcess(userID)
	if err != nil {
		return "", "", err
	}

	processID, err := process.Save(c, mongoSession)
	if err != nil {
		return "", "", err
	}

	return processID, process.Code, nil
}

// VerifyProcess : verify a process and returns its unique ID and its stored value.
func VerifyProcess(c *gin.Context, processID, key string) (string, string, error) {
	if _, err := bson.ObjectIDFromHex(processID); err != nil {
		return "", "", errors.WrongObjectID()
	}

	mongoSession := session.Get(c)

	process := &model.UpdatePropertyProcess{}
	if err := process.GetByID(c, processID, mongoSession); err != nil || !process.IsSaved() {
		return "", "", errors.NotFound()
	}

	// Do not validate or cancel if the key is not verified.
	if process.Code != key {
		if process.Tries >= 2 {
			process.Delete(c, mongoSession)
		} else {
			process.Increment(c, mongoSession)
		}
		return "", "", errors.NotFound()
	}

	return process.UserID.Hex(), process.Value, nil
}

// DeleteProcessByID : deletes a process by its unique ID.
func DeleteProcessByID(c *gin.Context, processID string) error {
	objectID, err := bson.ObjectIDFromHex(processID)
	if err != nil {
		return errors.WrongObjectID()
	}

	mongoSession := session.Get(c)

	process := &model.UpdatePropertyProcess{
		ID: objectID,
	}

	return process.Delete(c, mongoSession)
}
