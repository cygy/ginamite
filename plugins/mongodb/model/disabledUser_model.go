package model

import (
	"context"
	"time"

	"github.com/cygy/ginamite/common/errors"
	"github.com/cygy/ginamite/common/mongo/document"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// DisabledUser : properties of a disabled user.
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

// NewDisabledUser : returns a new 'DisabledUser' struct.
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

// IsSaved : returns true if the 'DisabledUser' document if saved in the collection.
func (user *DisabledUser) IsSaved() bool {
	return !user.ID.IsZero()
}

// Save : inserts the 'DisabledUser' document in the collection, returns an error if needed.
func (user *DisabledUser) Save(ctx context.Context, db *mongo.Database) (string, error) {
	return document.SaveDocument(ctx, user, DisabledUserCollection, db)
}

// Update : updates the 'DisabledUser' document, returns an error if needed.
func (user *DisabledUser) Update(ctx context.Context, update any, db *mongo.Database) error {
	return document.UpdateDocument(ctx, user.ID, DisabledUserCollection, update, db)
}

// Delete : deletes the 'DisabledUser' document from the collection, returns an error if needed.
func (user *DisabledUser) Delete(ctx context.Context, db *mongo.Database) error {
	return document.DeleteDocument(ctx, user.ID, DisabledUserCollection, db)
}

// GetByID : initializes the 'DisabledUser' struct from the 'DisabledUser' document with the ID.
func (user *DisabledUser) GetByID(ctx context.Context, id string, db *mongo.Database) error {
	return document.GetDocumentByID(ctx, user, DisabledUserCollection, id, db)
}

// GetID : returns the 'DisabledUser' ID.
func (user *DisabledUser) GetID() string {
	return user.ID.Hex()
}

// SetID : initializes the 'DisabledUser' ID.
func (user *DisabledUser) SetID(id string) error {
	objectID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return errors.WrongObjectID()
	}
	user.ID = objectID
	return nil
}
