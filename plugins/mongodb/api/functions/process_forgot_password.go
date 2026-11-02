package functions

import (
	"context"
	"errors"

	commonErrors "github.com/cygy/ginamite/common/errors"
	"github.com/cygy/ginamite/common/log"
	"github.com/cygy/ginamite/plugins/mongodb/model"
	"github.com/cygy/ginamite/plugins/mongodb/mongo/bson"
)

// NewForgotPasswordProcessForUser creates a forgot password process for the user, replacing the existing one, and
// returns its unique ID. The private key is the code to give to the user.
func (p *Processes) NewForgotPasswordProcessForUser(ctx context.Context, userID, processPrivateKey string) (string, error) {
	mongoSession := p.database()

	// Delete the existing forgot password process.
	if err := model.DeleteForgotPasswordProcessByUserID(ctx, userID, mongoSession); err != nil && !commonErrors.IsNotFound(err) {
		return "", err
	}

	// Create a new forgot password process.
	fpp, err := model.NewForgotPasswordProcess(userID)
	if err != nil {
		return "", err
	}
	fpp.Code = processPrivateKey
	fppID, err := fpp.Save(ctx, mongoSession)
	if err != nil {
		return "", err
	}

	return fppID, nil
}

// GetForgotPasswordProcessDetailsByID returns the ID of the user and the private key of a forgot password process.
func (p *Processes) GetForgotPasswordProcessDetailsByID(ctx context.Context, processID string) (userID, processPrivateKey string, err error) {
	if _, err := bson.ObjectIDFromHex(processID); err != nil {
		return "", "", errors.New("wrong forgot password process ID")
	}

	mongoSession := p.database()

	fpp := &model.ForgotPasswordProcess{}
	if err := fpp.GetByID(ctx, processID, mongoSession); err != nil || !fpp.IsSaved() {
		return "", "", errors.New("unable to find the forgot password process")
	}

	return fpp.UserID.Hex(), fpp.Code, nil
}

// WrongAccessToForgotPasswordProcess records a wrong access to a forgot password process: it counts as a try, and
// the process is deleted after three wrong tries.
func (p *Processes) WrongAccessToForgotPasswordProcess(ctx context.Context, processID string) {
	mongoSession := p.database()

	fpp := &model.ForgotPasswordProcess{}
	if err := fpp.GetByID(ctx, processID, mongoSession); err != nil || !fpp.IsSaved() {
		return
	}

	if fpp.Tries >= 2 {
		if err := fpp.Delete(ctx, mongoSession); err != nil {
			log.Error("unable to delete", "error", err)
		}
	} else {
		if err := fpp.Increment(ctx, mongoSession); err != nil {
			log.Error("unable to increment", "error", err)
		}
	}
}

// DeleteForgotPasswordProcess deletes a forgot password process.
func (p *Processes) DeleteForgotPasswordProcess(ctx context.Context, processID string) error {
	objectID, err := bson.ObjectIDFromHex(processID)
	if err != nil {
		return err
	}

	mongoSession := p.database()

	process := &model.ForgotPasswordProcess{
		ID: objectID,
	}

	return process.Delete(ctx, mongoSession)
}
