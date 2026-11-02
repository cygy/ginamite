package model

import (
	"context"
	stderrors "errors"
	"time"

	"github.com/cygy/ginamite/common/authentication"
	"github.com/cygy/ginamite/common/config"
	"github.com/cygy/ginamite/common/errors"
	"github.com/cygy/ginamite/plugins/mongodb/mongo/bson"
	"github.com/cygy/ginamite/plugins/mongodb/mongo/document"
	"github.com/cygy/ginamite/plugins/mongodb/mongo/store"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// UserCollection is the name of the collection of the user documents.
const UserCollection = "users"

// User is a user of the application.
type User struct {
	ID                     bson.ObjectID             `json:"id" bson:"_id,omitempty"`
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
	LastLogin              time.Time                 `json:"last_login,omitempty" bson:"lastLogin,omitempty"`
	VersionOfTermsAccepted string                    `json:"version_of_terms_accepted" bson:"versionOfTermsAccepted"`
	PrivateKey             string                    `json:"-" bson:"privateKey"`
}

// UserPublicInfos holds the public information of a user.
type UserPublicInfos struct {
	Description map[string]string `json:"description,omitempty" bson:"description,omitempty"`
	Facebook    string            `json:"facebook,omitempty" bson:"facebook,omitempty"`
	Twitter     string            `json:"twitter,omitempty" bson:"twitter,omitempty"`
	Instagram   string            `json:"instagram,omitempty" bson:"instagram,omitempty"`
	Image       Image             `json:"image,omitempty" bson:"image,omitempty"`
}

// UserPrivateInfos holds the private information of a user.
type UserPrivateInfos struct {
	FirstName    string `json:"first_name" bson:"firstName"`
	LastName     string `json:"last_name" bson:"lastName"`
	Email        string `json:"email" bson:"email"`
	IsEmailValid bool   `json:"is_email_valid" bson:"isEmailValid"`
}

// UserSettings holds the settings of a user.
type UserSettings struct {
	Locale   string `json:"locale" bson:"locale"`
	Timezone string `json:"timezone" bson:"timezone"`
}

// UserRegistrationInfos holds the details of the registration of a user.
type UserRegistrationInfos struct {
	Date         time.Time `json:"date" bson:"date"`
	Source       string    `json:"source" bson:"source"`
	Method       string    `json:"method" bson:"method"`
	Device       Device    `json:"device" bson:"device"`
	IPAddress    string    `json:"ip_address" bson:"ip"`
	Code         string    `json:"-" bson:"code"`
	CodeAttempts uint      `json:"-" bson:"codeAttempts"`
}

// UserIdentificationMethods holds the credentials of a user for each identification method.
type UserIdentificationMethods struct {
	Password string         `json:"-" bson:"password"`
	Facebook UserOAuthInfos `json:"facebook,omitempty" bson:"facebook,omitempty"`
	Google   UserOAuthInfos `json:"google,omitempty" bson:"google,omitempty"`
	Twitter  UserOAuthInfos `json:"twitter,omitempty" bson:"twitter,omitempty"`
}

// UserOAuthInfos holds the information of a user received from an OAuth service.
type UserOAuthInfos struct {
	UserID    string    `json:"user_id" bson:"userId"`
	Token     string    `json:"token" bson:"token"`
	FirstName string    `json:"first_name" bson:"firstName,omitempty"`
	LastName  string    `json:"last_name" bson:"lastName,omitempty"`
	Email     string    `json:"email" bson:"email,omitempty"`
	CreatedAt time.Time `json:"created_at" bson:"createdAt,omitempty"`
	ExpiresAt time.Time `json:"expires_at" bson:"expiresAt,omitempty"`
}

// UserNotificationsStats holds the counts of the notifications received by a user.
type UserNotificationsStats struct {
	Total  int64 `json:"total" bson:"total"`
	Unread int   `json:"unread" bson:"unread"`
}

// NewUser returns a new User.
func NewUser() *User {
	return &User{
		IsAnonymous: false,
		IsAdmin:     false,
	}
}

// UserWithID returns a new User with the ID, or an error if the ID is not a valid ObjectID.
func UserWithID(id string) (*User, error) {
	user := NewUser()

	objectID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	user.ID = objectID
	return user, nil
}

// NewUserFromDisabledUser returns the User to restore from a DisabledUser.
func NewUserFromDisabledUser(user *DisabledUser) *User {
	return &User{
		ID:                     user.UserID,
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
	}
}

// IsSaved reports whether the user document is saved in the collection.
func (user *User) IsSaved() bool {
	return bson.IsObjectIDProvided(user.ID)
}

// Save inserts the user document in the collection, returning an error if needed.
func (user *User) Save(ctx context.Context, db store.Database) (string, error) {
	return document.SaveDocument(ctx, user, UserCollection, db)
}

// Update updates the user document, returning an error if needed.
func (user *User) Update(ctx context.Context, update any, db store.Database) error {
	return document.UpdateDocument(ctx, user.ID, UserCollection, update, db)
}

// Delete deletes the user document from the collection, returning an error if needed.
// The documents linked to the user are not deleted here: see the worker functions.
func (user *User) Delete(ctx context.Context, db store.Database) error {
	return document.DeleteDocument(ctx, user.ID, UserCollection, db)
}

// GetByID loads the User document with the ID.
func (user *User) GetByID(ctx context.Context, id string, db store.Database) error {
	return document.GetDocumentByID(ctx, user, UserCollection, id, db)
}

// GetID returns the User ID.
func (user *User) GetID() string {
	return user.ID.Hex()
}

// SetID sets the ID of the User from its hexadecimal form. The ID is kept if the form is not valid.
func (user *User) SetID(id string) error {
	objectID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	user.ID = objectID
	return nil
}

// SetAdmin sets or unsets the administrator flag of the user. It is meant for the tests: it does nothing
// outside of the test environment of the configuration. A nil configuration is the test environment.
func (user *User) SetAdmin(ctx context.Context, configuration *config.Configuration, admin bool, db store.Database) error {
	if configuration == nil || configuration.IsTestEnvironment() {
		return user.Update(ctx, bson.M{"$set": bson.M{"isAdmin": admin}}, db)
	}

	return nil
}

// GetByUsername loads the User document with the username.
func (user *User) GetByUsername(ctx context.Context, username string, db store.Database) error {
	return document.GetDocumentBySelector(ctx, user, UserCollection, bson.M{"username": username}, db)
}

// GetByEmailAddress loads the User document with the email address.
func (user *User) GetByEmailAddress(ctx context.Context, emailAddress string, db store.Database) error {
	return document.GetDocumentBySelector(ctx, user, UserCollection, bson.M{"privateInfos.email": emailAddress}, db)
}

// GetByFacebookUserID loads the User document with the facebook userId.
func (user *User) GetByFacebookUserID(ctx context.Context, userID string, db store.Database) error {
	return document.GetDocumentBySelector(ctx, user, UserCollection, bson.M{"identificationMethods.facebook.userId": userID}, db)
}

// GetByGoogleUserID loads the User document with the google userId.
func (user *User) GetByGoogleUserID(ctx context.Context, userID string, db store.Database) error {
	return document.GetDocumentBySelector(ctx, user, UserCollection, bson.M{"identificationMethods.google.userId": userID}, db)
}

// GetByIDAndNotification loads the User document with the email address and a notification type.
func (user *User) GetByIDAndNotification(ctx context.Context, userID, notificationType string, db store.Database) error {
	objectID, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		return err
	}
	return document.GetDocumentBySelector(ctx, user, UserCollection, bson.D{
		{Key: "_id", Value: objectID},
		{Key: "notifications", Value: bson.M{"$in": []string{notificationType}}},
	}, db)
}

// GetByIdentifier loads the user document with the username or, if there is none, the
// email address. It returns a "not found" error, see errors.IsNotFound, if neither matches.
func (user *User) GetByIdentifier(ctx context.Context, identifier string, db store.Database) error {
	if err := document.GetDocumentBySelector(ctx, user, UserCollection, bson.D{{Key: "username", Value: identifier}}, db); err != nil && !IsNotFound(err) {
		return err
	}

	if !user.IsSaved() {
		if err := document.GetDocumentBySelector(ctx, user, UserCollection, bson.D{{Key: "privateInfos.email", Value: identifier}}, db); err != nil && !IsNotFound(err) {
			return err
		}
	}

	if !user.IsSaved() {
		return errors.NotFound()
	}

	return nil
}

// UpdateLastLogin updates the last login date.
func (user *User) UpdateLastLogin(ctx context.Context, db store.Database) error {
	return user.Update(ctx, bson.M{"$set": bson.M{"lastLogin": time.Now()}}, db)
}

// ToUser returns the authentication.User of the user.
func (user *User) ToUser() authentication.User {
	return authentication.User{
		ID:           user.ID.Hex(),
		Username:     user.Username,
		EmailAddress: user.PrivateInfos.Email,
		Locale:       user.Settings.Locale,
		PrivateKey:   user.PrivateKey,
		IsEmailValid: user.PrivateInfos.IsEmailValid,
	}
}

// IsUserWithUsernameExisting reports whether a user with this username already exists.
func IsUserWithUsernameExisting(ctx context.Context, username string, db store.Database) (bool, error) {
	count, err := document.CountBySelector(ctx, UserCollection, bson.M{"username": username}, db)
	return count > 0, err
}

// IsUserWithEmailAddressExisting reports whether a user with this email address already exists.
func IsUserWithEmailAddressExisting(ctx context.Context, emailAddress string, db store.Database) (bool, error) {
	count, err := document.CountBySelector(ctx, UserCollection, bson.M{"privateInfos.email": emailAddress}, db)
	return count > 0, err
}

// IsUserWithFacebookUserIDExisting reports whether a user with this facebook user id already exists.
func IsUserWithFacebookUserIDExisting(ctx context.Context, userID string, db store.Database) (bool, error) {
	count, err := document.CountBySelector(ctx, UserCollection, bson.M{"identificationMethods.facebook.userId": userID}, db)
	return count > 0, err
}

// IsUserWithGoogleUserIDExisting reports whether a user with this google user id already exists.
func IsUserWithGoogleUserIDExisting(ctx context.Context, userID string, db store.Database) (bool, error) {
	count, err := document.CountBySelector(ctx, UserCollection, bson.M{"identificationMethods.google.userId": userID}, db)
	return count > 0, err
}

// IsUserWithValidEmailAddressExisting reports whether a user has this email address, and it is valid.
func IsUserWithValidEmailAddressExisting(ctx context.Context, emailAddress string, db store.Database) (bool, error) {
	count, err := document.CountBySelector(ctx, UserCollection, bson.M{"privateInfos.email": emailAddress, "privateInfos.isEmailValid": true}, db)
	return count > 0, err
}

// DeleteUserByUserID deletes the document with the ID.
func DeleteUserByUserID(ctx context.Context, userID string, db store.Database) error {
	objectID, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		return err
	}
	return document.DeleteDocument(ctx, objectID, UserCollection, db)
}

// GetAnonymousUserID returns the ID of the anonymous user.
func GetAnonymousUserID(ctx context.Context, db store.Database) (string, error) {
	user := User{}
	err := document.GetDocumentBySelector(ctx, &user, UserCollection, bson.M{"isAnonymous": true}, db)
	return user.ID.Hex(), err
}

// GetAnonymousUser returns the anonymous user.
func GetAnonymousUser(ctx context.Context, db store.Database) (*User, error) {
	user := &User{}
	err := document.GetDocumentBySelector(ctx, user, UserCollection, bson.M{"isAnonymous": true}, db)
	return user, err
}

// GetNeverUsedUserIDs returns the IDs of the non-admin users who registered more than interval ago and never signed in.
func GetNeverUsedUserIDs(ctx context.Context, interval time.Duration, db store.Database) ([]string, error) {
	users := []User{}
	err := document.GetDocumentsBySelector(ctx, &users, UserCollection,
		bson.D{
			{Key: "registrationInfos.date", Value: bson.M{"$lt": time.Now().Add(interval * -1)}},
			{Key: "lastLogin", Value: bson.M{"$exists": false}},
			{Key: "isAdmin", Value: false},
		},
		db)

	userIDs := make([]string, len(users))
	for i, user := range users {
		userIDs[i] = user.ID.Hex()
	}

	return userIDs, err
}

// GetInactiveUserIDs returns the IDs of the users whose last sign-in is older than interval.
func GetInactiveUserIDs(ctx context.Context, interval time.Duration, db store.Database) ([]string, error) {
	users := []User{}
	err := document.GetDocumentsBySelector(ctx, &users, UserCollection, bson.D{{Key: "lastLogin", Value: bson.M{"$lt": time.Now().Add(interval * -1)}}}, db)

	userIDs := make([]string, len(users))
	for i, user := range users {
		userIDs[i] = user.ID.Hex()
	}

	return userIDs, err
}

// IsNotFound reports whether the error means that no document was found, whether it comes from the driver or from
// document.GetDocumentByID.
func IsNotFound(err error) bool {
	return stderrors.Is(err, mongo.ErrNoDocuments) || errors.IsNotFound(err)
}
