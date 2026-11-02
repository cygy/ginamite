package functions

import (
	"context"
	"crypto/subtle"

	"github.com/cygy/ginamite/common/errors"
	"github.com/cygy/ginamite/common/log"
	"github.com/cygy/ginamite/plugins/mongodb/model"
	"github.com/cygy/ginamite/plugins/mongodb/mongo/bson"
	"github.com/cygy/ginamite/plugins/mongodb/mongo/store"
)

// Processes are the functions of the processes bound to a user, such as the forgot password ones, working on a MongoDB
// database.
type Processes struct {
	database func() store.Database
}

// NewProcesses returns the functions of the processes working on the database returned by the function, which is called
// each time a function needs the database.
func NewProcesses(database func() store.Database) *Processes {
	return &Processes{database: database}
}

// CreateNewDeleteAccountProcess creates a process to delete the account of the user, replacing the existing one.
// It returns the unique ID of the process and its code.
func (p *Processes) CreateNewDeleteAccountProcess(ctx context.Context, userID string) (string, string, error) {
	return p.createProcess(ctx, userID, model.DeleteDeleteAccountProcessByUserID, func() (*model.UpdatePropertyProcess, error) {
		return model.NewDeleteAccountProcess(userID)
	})
}

// CreateNewDisableAccountProcess creates a process to disable the account of the user, replacing the existing one.
// It returns the unique ID of the process and its code.
func (p *Processes) CreateNewDisableAccountProcess(ctx context.Context, userID string) (string, string, error) {
	return p.createProcess(ctx, userID, model.DeleteDisableAccountProcessByUserID, func() (*model.UpdatePropertyProcess, error) {
		return model.NewDisableAccountProcess(userID)
	})
}

// CreateNewUpdatePasswordProcess creates a process to update the password of the user, replacing the existing one.
// It returns the unique ID of the process and its code.
func (p *Processes) CreateNewUpdatePasswordProcess(ctx context.Context, userID, encryptedPassword string) (string, string, error) {
	return p.createProcess(ctx, userID, model.DeleteUpdatePasswordProcessByUserID, func() (*model.UpdatePropertyProcess, error) {
		return model.NewUpdatePasswordProcess(userID, encryptedPassword)
	})
}

// CreateNewUpdateEmailAddressProcess creates a process to update the email address of the user, replacing the
// existing one. It returns the unique ID of the process and its code.
func (p *Processes) CreateNewUpdateEmailAddressProcess(ctx context.Context, userID, emailAddress string) (string, string, error) {
	return p.createProcess(ctx, userID, model.DeleteUpdateEmailAddressProcessByUserID, func() (*model.UpdatePropertyProcess, error) {
		return model.NewUpdateEmailAddressProcess(userID, emailAddress)
	})
}

// CreateNewVerifyEmailAddressProcess creates a process to verify the email address of the user, replacing the
// existing one. It returns the unique ID of the process and its code.
func (p *Processes) CreateNewVerifyEmailAddressProcess(ctx context.Context, userID string) (string, string, error) {
	return p.createProcess(ctx, userID, model.DeleteUpdateValidEmailAddressProcessByUserID, func() (*model.UpdatePropertyProcess, error) {
		return model.NewUpdateValidEmailAddressProcess(userID)
	})
}

// createProcess deletes the existing process of the user, saves the new one and returns its ID and its code.
func (p *Processes) createProcess(ctx context.Context, userID string, deleteExisting func(context.Context, string, store.Database) error, newProcess func() (*model.UpdatePropertyProcess, error)) (string, string, error) {
	mongoSession := p.database()

	if err := deleteExisting(ctx, userID, mongoSession); err != nil && !errors.IsNotFound(err) {
		return "", "", err
	}

	process, err := newProcess()
	if err != nil {
		return "", "", err
	}

	processID, err := process.Save(ctx, mongoSession)
	if err != nil {
		return "", "", err
	}

	return processID, process.Code, nil
}

// VerifyProcess verifies the key of a process and returns the ID of its user and its stored value.
// A wrong key counts as a try: the process is deleted after three wrong tries. It returns a "not found" error if the
// process does not exist or if the key is wrong.
func (p *Processes) VerifyProcess(ctx context.Context, processID, key string) (string, string, error) {
	if _, err := bson.ObjectIDFromHex(processID); err != nil {
		return "", "", err
	}

	mongoSession := p.database()

	process := &model.UpdatePropertyProcess{}
	if err := process.GetByID(ctx, processID, mongoSession); err != nil || !process.IsSaved() {
		return "", "", errors.ErrNotFound
	}

	// Do not validate or cancel if the key is not verified.
	if subtle.ConstantTimeCompare([]byte(process.Code), []byte(key)) != 1 {
		if process.Tries >= 2 {
			if err := process.Delete(ctx, mongoSession); err != nil {
				log.Error("unable to delete", "error", err)
			}
		} else {
			if err := process.Increment(ctx, mongoSession); err != nil {
				log.Error("unable to increment", "error", err)
			}
		}
		return "", "", errors.ErrNotFound
	}

	return process.UserID.Hex(), process.Value, nil
}

// DeleteProcessByID deletes a process by its unique ID.
func (p *Processes) DeleteProcessByID(ctx context.Context, processID string) error {
	objectID, err := bson.ObjectIDFromHex(processID)
	if err != nil {
		return err
	}

	mongoSession := p.database()

	process := &model.UpdatePropertyProcess{
		ID: objectID,
	}

	return process.Delete(ctx, mongoSession)
}
