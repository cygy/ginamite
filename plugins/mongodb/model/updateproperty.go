package model

import (
	"context"
	"fmt"

	"github.com/cygy/ginamite/common/authentication"
	"github.com/cygy/ginamite/common/random"
	"github.com/cygy/ginamite/plugins/mongodb/mongo/bson"
	"github.com/cygy/ginamite/plugins/mongodb/mongo/document"
	"github.com/cygy/ginamite/plugins/mongodb/mongo/store"
)

// UpdatePropertyProcessCollection is the name of the collection of the updateProperty documents.
const UpdatePropertyProcessCollection = "updateProperty"

// UpdatePropertyProcess is a pending update of a property of a user, to confirm with a code.
type UpdatePropertyProcess struct {
	ID     bson.ObjectID `json:"id" bson:"_id,omitempty"`
	UserID bson.ObjectID `json:"user_id" bson:"userId"`
	Type   string        `json:"type" bson:"type"`
	Value  string        `json:"value" bson:"value"`
	Code   string        `json:"code" bson:"code"`
	Tries  uint          `json:"-" bson:"tries"`
	TTLKey string        `json:"-" bson:"userIdtype"` // This field is indexed with a TTL. A TTL index must use only one field, so the pair 'userId / type' can not be used.
}

// NewUpdatePropertyProcess returns a new UpdatePropertyProcess.
func NewUpdatePropertyProcess(userID, t, value string) (*UpdatePropertyProcess, error) {
	objectID, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		return nil, err
	}

	return &UpdatePropertyProcess{
		UserID: objectID,
		Type:   t,
		Value:  value,
		Code:   random.String(64),
		Tries:  0,
		TTLKey: fmt.Sprintf("%s%s", userID, t),
	}, nil
}

// NewUpdatePasswordProcess returns a new UpdatePropertyProcess.
func NewUpdatePasswordProcess(userID, value string) (*UpdatePropertyProcess, error) {
	return NewUpdatePropertyProcess(userID, authentication.ProcessUpdatePassword, value)
}

// NewUpdateEmailAddressProcess returns a new UpdatePropertyProcess.
func NewUpdateEmailAddressProcess(userID, value string) (*UpdatePropertyProcess, error) {
	return NewUpdatePropertyProcess(userID, authentication.ProcessUpdateEmailAddress, value)
}

// NewUpdateValidEmailAddressProcess returns a new UpdatePropertyProcess.
func NewUpdateValidEmailAddressProcess(userID string) (*UpdatePropertyProcess, error) {
	return NewUpdatePropertyProcess(userID, authentication.ProcessVerifyEmailAddress, "true")
}

// NewDeleteAccountProcess returns a new UpdatePropertyProcess.
func NewDeleteAccountProcess(userID string) (*UpdatePropertyProcess, error) {
	return NewUpdatePropertyProcess(userID, authentication.ProcessDeleteAccount, "true")
}

// NewDisableAccountProcess returns a new UpdatePropertyProcess.
func NewDisableAccountProcess(userID string) (*UpdatePropertyProcess, error) {
	return NewUpdatePropertyProcess(userID, authentication.ProcessDisableAccount, "true")
}

// IsSaved reports whether the UpdatePropertyProcess document is saved in the collection.
func (process *UpdatePropertyProcess) IsSaved() bool {
	return bson.IsObjectIDProvided(process.ID)
}

// Save inserts the UpdatePropertyProcess document in the collection, returning an error if needed.
func (process *UpdatePropertyProcess) Save(ctx context.Context, db store.Database) (string, error) {
	return document.SaveDocument(ctx, process, UpdatePropertyProcessCollection, db)
}

// Update updates the UpdatePropertyProcess document, returning an error if needed.
func (process *UpdatePropertyProcess) Update(ctx context.Context, update any, db store.Database) error {
	return document.UpdateDocument(ctx, process.ID, UpdatePropertyProcessCollection, update, db)
}

// Delete deletes the UpdatePropertyProcess document from the collection, returning an error if needed.
func (process *UpdatePropertyProcess) Delete(ctx context.Context, db store.Database) error {
	return document.DeleteDocument(ctx, process.ID, UpdatePropertyProcessCollection, db)
}

// GetByID loads the UpdatePropertyProcess document with the ID.
func (process *UpdatePropertyProcess) GetByID(ctx context.Context, id string, db store.Database) error {
	return document.GetDocumentByID(ctx, process, UpdatePropertyProcessCollection, id, db)
}

// GetID returns the UpdatePropertyProcess ID.
func (process *UpdatePropertyProcess) GetID() string {
	return process.ID.Hex()
}

// SetID sets the ID of the UpdatePropertyProcess from its hexadecimal form. The ID is kept if the form is not valid.
func (process *UpdatePropertyProcess) SetID(id string) error {
	objectID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	process.ID = objectID
	return nil
}

// GetByUserIDAndType loads the process of the user for the type.
func (process *UpdatePropertyProcess) GetByUserIDAndType(ctx context.Context, userID, t string, db store.Database) error {
	objectID, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		return err
	}
	return document.GetDocumentBySelector(ctx, process, UpdatePropertyProcessCollection, bson.D{
		{Key: "userId", Value: objectID},
		{Key: "type", Value: t},
	}, db)
}

// GetUpdatePasswordProcessByUserID returns a UpdatePropertyProcess struct with a user.
func (process *UpdatePropertyProcess) GetUpdatePasswordProcessByUserID(ctx context.Context, userID string, db store.Database) error {
	return process.GetByUserIDAndType(ctx, userID, authentication.ProcessUpdatePassword, db)
}

// GetUpdateEmailAddressProcessByUserID returns a UpdatePropertyProcess struct with a user.
func (process *UpdatePropertyProcess) GetUpdateEmailAddressProcessByUserID(ctx context.Context, userID string, db store.Database) error {
	return process.GetByUserIDAndType(ctx, userID, authentication.ProcessUpdateEmailAddress, db)
}

// GetUpdateValidEmailAddressProcessByUserID returns a UpdatePropertyProcess struct with a user.
func (process *UpdatePropertyProcess) GetUpdateValidEmailAddressProcessByUserID(ctx context.Context, userID string, db store.Database) error {
	return process.GetByUserIDAndType(ctx, userID, authentication.ProcessVerifyEmailAddress, db)
}

// GetDeleteAccountProcessByUserID returns a UpdatePropertyProcess struct with a user.
func (process *UpdatePropertyProcess) GetDeleteAccountProcessByUserID(ctx context.Context, userID string, db store.Database) error {
	return process.GetByUserIDAndType(ctx, userID, authentication.ProcessDeleteAccount, db)
}

// GetDisableAccountProcessByUserID returns a UpdatePropertyProcess struct with a user.
func (process *UpdatePropertyProcess) GetDisableAccountProcessByUserID(ctx context.Context, userID string, db store.Database) error {
	return process.GetByUserIDAndType(ctx, userID, authentication.ProcessDisableAccount, db)
}

// Increment increments the number of tries of the process.
func (process *UpdatePropertyProcess) Increment(ctx context.Context, db store.Database) error {
	return process.Update(ctx, bson.M{"$inc": bson.M{"tries": 1}}, db)
}

// DeleteUpdatePropertyProcessByUserIDAndType deletes a process linked to a user.
func DeleteUpdatePropertyProcessByUserIDAndType(ctx context.Context, userID, t string, db store.Database) error {
	objectID, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		return err
	}
	_, err = document.DeleteDocumentsBySelector(ctx, UpdatePropertyProcessCollection, bson.D{
		{Key: "userId", Value: objectID},
		{Key: "type", Value: t},
	}, db)
	return err
}

// DeleteUpdatePasswordProcessByUserID deletes a process linked to a user.
func DeleteUpdatePasswordProcessByUserID(ctx context.Context, userID string, db store.Database) error {
	return DeleteUpdatePropertyProcessByUserIDAndType(ctx, userID, authentication.ProcessUpdatePassword, db)
}

// DeleteUpdateEmailAddressProcessByUserID deletes a process linked to a user.
func DeleteUpdateEmailAddressProcessByUserID(ctx context.Context, userID string, db store.Database) error {
	return DeleteUpdatePropertyProcessByUserIDAndType(ctx, userID, authentication.ProcessUpdateEmailAddress, db)
}

// DeleteUpdateValidEmailAddressProcessByUserID deletes a process linked to a user.
func DeleteUpdateValidEmailAddressProcessByUserID(ctx context.Context, userID string, db store.Database) error {
	return DeleteUpdatePropertyProcessByUserIDAndType(ctx, userID, authentication.ProcessVerifyEmailAddress, db)
}

// DeleteDeleteAccountProcessByUserID deletes a process linked to a user.
func DeleteDeleteAccountProcessByUserID(ctx context.Context, userID string, db store.Database) error {
	return DeleteUpdatePropertyProcessByUserIDAndType(ctx, userID, authentication.ProcessDeleteAccount, db)
}

// DeleteDisableAccountProcessByUserID deletes a process linked to a user.
func DeleteDisableAccountProcessByUserID(ctx context.Context, userID string, db store.Database) error {
	return DeleteUpdatePropertyProcessByUserIDAndType(ctx, userID, authentication.ProcessDisableAccount, db)
}
