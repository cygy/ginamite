package functions

import (
	"errors"

	commonErrors "github.com/cygy/ginamite/common/errors"
	"github.com/cygy/ginamite/common/mongo/session"
	"github.com/cygy/ginamite/plugins/mongodb/model"

	"github.com/gin-gonic/gin"
	"github.com/globalsign/mgo/bson"
)

// CreateNewDeleteAccountProcess : creates a process and returns its unique ID.
func CreateNewDeleteAccountProcess(c *gin.Context, userID string) (string, string, error) {
	mongoSession := session.Get(c)

	// Delete the existing process.
	if err := model.DeleteDeleteAccountProcessByUserID(userID, mongoSession); err != nil && !commonErrors.IsNotFound(err) {
		return "", "", err
	}

	// Create a new process.
	process := model.NewDeleteAccountProcess(userID)
	processID, err := process.Save(mongoSession)
	if err != nil {
		return "", "", err
	}

	return processID, process.Code, nil
}

// CreateNewDisableAccountProcess : creates a process and returns its unique ID.
func CreateNewDisableAccountProcess(c *gin.Context, userID string) (string, string, error) {
	mongoSession := session.Get(c)

	// Delete the existing process.
	if err := model.DeleteDisableAccountProcessByUserID(userID, mongoSession); err != nil && !commonErrors.IsNotFound(err) {
		return "", "", err
	}

	// Create a new process.
	process := model.NewDisableAccountProcess(userID)
	processID, err := process.Save(mongoSession)
	if err != nil {
		return "", "", err
	}

	return processID, process.Code, nil
}

// CreateNewUpdatePasswordProcess : creates a process and returns its unique ID.
func CreateNewUpdatePasswordProcess(c *gin.Context, userID, encryptedPassword string) (string, string, error) {
	mongoSession := session.Get(c)

	// Delete the existing process.
	if err := model.DeleteUpdatePasswordProcessByUserID(userID, mongoSession); err != nil && !commonErrors.IsNotFound(err) {
		return "", "", err
	}

	// Create a new process.
	process := model.NewUpdatePasswordProcess(userID, encryptedPassword)
	processID, err := process.Save(mongoSession)
	if err != nil {
		return "", "", err
	}

	return processID, process.Code, nil
}

// CreateNewUpdateEmailAddressProcess : creates a process and returns its unique ID.
func CreateNewUpdateEmailAddressProcess(c *gin.Context, userID, emailAddress string) (string, string, error) {
	mongoSession := session.Get(c)

	// Delete the existing process.
	if err := model.DeleteUpdateEmailAddressProcessByUserID(userID, mongoSession); err != nil && !commonErrors.IsNotFound(err) {
		return "", "", err
	}

	// Create a new process.
	process := model.NewUpdateEmailAddressProcess(userID, emailAddress)
	processID, err := process.Save(mongoSession)
	if err != nil {
		return "", "", err
	}

	return processID, process.Code, nil
}

// CreateNewVerifyEmailAddressProcess : creates a process and returns its unique ID.
func CreateNewVerifyEmailAddressProcess(c *gin.Context, userID string) (string, string, error) {
	mongoSession := session.Get(c)

	// Delete the existing process.
	if err := model.DeleteUpdateValidEmailAddressProcessByUserID(userID, mongoSession); err != nil && !commonErrors.IsNotFound(err) {
		return "", "", err
	}

	// Create a new process.
	process := model.NewUpdateValidEmailAddressProcess(userID)
	processID, err := process.Save(mongoSession)
	if err != nil {
		return "", "", err
	}

	return processID, process.Code, nil
}

// VerifyProcess : verify a process and returns its unique ID and its stored value.
func VerifyProcess(c *gin.Context, processID, key string) (string, string, error) {
	if !bson.IsObjectIdHex(processID) {
		return "", "", errors.New("wrong object ID")
	}

	mongoSession := session.Get(c)

	process := &model.UpdatePropertyProcess{}
	if err := process.GetByID(processID, mongoSession); err != nil || !process.IsSaved() {
		return "", "", commonErrors.NotFound()
	}

	// Do not validate or cancel if the key is not verified.
	if process.Code != key {
		if process.Tries >= 2 {
			process.Delete(mongoSession)
		} else {
			process.Increment(mongoSession)
		}
		return "", "", commonErrors.NotFound()
	}

	return process.UserID.Hex(), process.Value, nil
}

// DeleteProcessByID : deletes a process by its unique ID.
func DeleteProcessByID(c *gin.Context, processID string) error {
	mongoSession := session.Get(c)

	process := &model.UpdatePropertyProcess{
		ID: bson.ObjectIdHex(processID),
	}

	return process.Delete(mongoSession)
}
