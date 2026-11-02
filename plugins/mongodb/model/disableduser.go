package model

import (
	"context"
	"time"

	"github.com/cygy/ginamite/plugins/mongodb/mongo/bson"
	"github.com/cygy/ginamite/plugins/mongodb/mongo/document"
	"github.com/cygy/ginamite/plugins/mongodb/mongo/store"
)

// DisabledUserCollection is the name of the collection of the disabledUsers documents.
const DisabledUserCollection = "disabledUsers"

// DisabledUser is a user that disabled their account, with the data to restore it.
type DisabledUser struct {
	ID                     bson.ObjectID             `json:"id" bson:"_id,omitempty"`
	UserID                 bson.ObjectID             `json:"user_id" bson:"userId"`
	Username               string                    `json:"username" bson:"username"`
	Slug                   string                    `json:"slug" bson:"slug"`
	Abilities              []string                  `json:"abilities,omitempty" bson:"abilities,omitempty"`
	Notifications          []string                  `json:"notifications" bson:"notifications"`
	NotificationsStats     UserNotificationsStats    `json:"notifications_stats" bson:"notificationsStats"`
	IdentificationMethods  UserIdentificationMethods `json:"identification_methods,omitempty" bson:"identificationMethods,omitempty"`
	PublicInfos            UserPublicInfos           `json:"public_infos" bson:"publicInfos,omitempty"`
	PrivateInfos           UserPrivateInfos          `json:"private_infos" bson:"privateInfos,omitempty"`
	RegistrationInfos      UserRegistrationInfos     `json:"registration_infos,omitempty" bson:"registrationInfos,omitempty"`
	Settings               UserSettings              `json:"settings" bson:"settings"`
	IsAdmin                bool                      `json:"is_admin" bson:"isAdmin"`
	IsAnonymous            bool                      `json:"is_anonymous" bson:"isAnonymous"`
	Verified               bool                      `json:"verified" bson:"verified"`
	VersionOfTermsAccepted string                    `json:"version_of_terms_accepted" bson:"versionOfTermsAccepted"`
	PrivateKey             string                    `json:"-" bson:"privateKey"`
	DisabledAt             time.Time                 `json:"disabled_at" bson:"disabledAt"`
}

// NewDisabledUser returns a new DisabledUser.
func NewDisabledUser(user *User) *DisabledUser {
	return &DisabledUser{
		UserID:                 user.ID,
		Username:               user.Username,
		Slug:                   user.Slug,
		Abilities:              user.Abilities,
		Notifications:          user.Notifications,
		NotificationsStats:     user.NotificationsStats,
		IdentificationMethods:  user.IdentificationMethods,
		PublicInfos:            user.PublicInfos,
		PrivateInfos:           user.PrivateInfos,
		RegistrationInfos:      user.RegistrationInfos,
		Settings:               user.Settings,
		IsAdmin:                user.IsAdmin,
		IsAnonymous:            user.IsAnonymous,
		Verified:               user.Verified,
		VersionOfTermsAccepted: user.VersionOfTermsAccepted,
		PrivateKey:             user.PrivateKey,
		DisabledAt:             time.Now(),
	}
}

// IsSaved reports whether the DisabledUser document is saved in the collection.
func (user *DisabledUser) IsSaved() bool {
	return bson.IsObjectIDProvided(user.ID)
}

// Save inserts the DisabledUser document in the collection, returning an error if needed.
func (user *DisabledUser) Save(ctx context.Context, db store.Database) (string, error) {
	return document.SaveDocument(ctx, user, DisabledUserCollection, db)
}

// Update updates the DisabledUser document, returning an error if needed.
func (user *DisabledUser) Update(ctx context.Context, update any, db store.Database) error {
	return document.UpdateDocument(ctx, user.ID, DisabledUserCollection, update, db)
}

// Delete deletes the DisabledUser document from the collection, returning an error if needed.
func (user *DisabledUser) Delete(ctx context.Context, db store.Database) error {
	return document.DeleteDocument(ctx, user.ID, DisabledUserCollection, db)
}

// GetByID loads the DisabledUser document with the ID.
func (user *DisabledUser) GetByID(ctx context.Context, id string, db store.Database) error {
	return document.GetDocumentByID(ctx, user, DisabledUserCollection, id, db)
}

// GetID returns the DisabledUser ID.
func (user *DisabledUser) GetID() string {
	return user.ID.Hex()
}

// SetID sets the ID of the DisabledUser from its hexadecimal form. The ID is kept if the form is not valid.
func (user *DisabledUser) SetID(id string) error {
	objectID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	user.ID = objectID
	return nil
}

// IsDisabledUserWithUsernameExisting reports whether a disabled user with this username already exists.
func IsDisabledUserWithUsernameExisting(ctx context.Context, username string, db store.Database) (bool, error) {
	count, err := document.CountBySelector(ctx, DisabledUserCollection, bson.M{"username": username}, db)
	return count > 0, err
}

// IsDisabledUserWithEmailAddressExisting reports whether a disabled user with this email address already exists.
func IsDisabledUserWithEmailAddressExisting(ctx context.Context, emailAddress string, db store.Database) (bool, error) {
	count, err := document.CountBySelector(ctx, DisabledUserCollection, bson.M{"privateInfos.email": emailAddress}, db)
	return count > 0, err
}

// IsDisabledUserWithFacebookUserIDExisting reports whether a disabled user with this facebook user id already exists.
func IsDisabledUserWithFacebookUserIDExisting(ctx context.Context, userID string, db store.Database) (bool, error) {
	count, err := document.CountBySelector(ctx, DisabledUserCollection, bson.M{"identificationMethods.facebook.userId": userID}, db)
	return count > 0, err
}

// IsDisabledUserWithGoogleUserIDExisting reports whether a disabled user with this google user id already exists.
func IsDisabledUserWithGoogleUserIDExisting(ctx context.Context, userID string, db store.Database) (bool, error) {
	count, err := document.CountBySelector(ctx, DisabledUserCollection, bson.M{"identificationMethods.google.userId": userID}, db)
	return count > 0, err
}

// IsDisabledUserWithValidEmailAddressExisting reports whether a disabled user has this email address, and it is valid.
func IsDisabledUserWithValidEmailAddressExisting(ctx context.Context, emailAddress string, db store.Database) (bool, error) {
	count, err := document.CountBySelector(ctx, DisabledUserCollection, bson.M{"privateInfos.email": emailAddress, "privateInfos.isEmailValid": true}, db)
	return count > 0, err
}

// DeleteDisabledUserByUserID deletes the document with the ID.
func DeleteDisabledUserByUserID(ctx context.Context, userID string, db store.Database) error {
	objectID, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		return err
	}
	_, err = document.DeleteDocumentsBySelector(ctx, DisabledUserCollection, bson.M{"userId": objectID}, db)
	return err
}

// GetDisabledUsersForAdmin returns the disabled users for the admin.
func GetDisabledUsersForAdmin(ctx context.Context, db store.Database) ([]AdminDisabledUser, error) {
	users := []AdminDisabledUser{}
	err := document.GetDocumentsBySort(ctx, &users, DisabledUserCollection, []string{"username"}, db)

	return users, err
}
