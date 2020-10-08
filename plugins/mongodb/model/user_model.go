package model

import (
	"time"

	"github.com/cygy/ginamite/common/config"
	"github.com/cygy/ginamite/common/mongo/document"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

// User : properties of a user.
type User struct {
	ID                     bson.ObjectId             `json:"id" bson:"_id,omitempty"`
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

// UserPublicInfos : public informations of a user.
type UserPublicInfos struct {
	Description map[string]string `json:"description,omitempty" bson:"description,omitempty"`
	Facebook    string            `json:"facebook,omitempty" bson:"facebook,omitempty"`
	Twitter     string            `json:"twitter,omitempty" bson:"twitter,omitempty"`
	Instagram   string            `json:"instagram,omitempty" bson:"instagram,omitempty"`
	Image       Image             `json:"image,omitempty" bson:"image,omitempty"`
}

// UserPrivateInfos : private informations of a user.
type UserPrivateInfos struct {
	FirstName    string `json:"first_name" bson:"firstName"`
	LastName     string `json:"last_name" bson:"lastName"`
	Email        string `json:"email" bson:"email"`
	IsEmailValid bool   `json:"is_email_valid" bson:"isEmailValid"`
}

// UserSettings : settings of a user.
type UserSettings struct {
	Locale   string `json:"locale" bson:"locale"`
	Timezone string `json:"timezone" bson:"timezone"`
}

// UserRegistrationInfos : private informations of a user.
type UserRegistrationInfos struct {
	Date         time.Time `json:"date" bson:"date"`
	Source       string    `json:"source" bson:"source"`
	Method       string    `json:"method" bson:"method"`
	Device       Device    `json:"device" bson:"device"`
	IPAddress    string    `json:"ip_address" bson:"ip"`
	Code         string    `json:"-" bson:"code"`
	CodeAttempts uint      `json:"-" bson:"codeAttempts"`
}

// UserIdentificationMethods : identification methods of a user.
type UserIdentificationMethods struct {
	Password string         `json:"-" bson:"password"`
	Facebook UserOAuthInfos `json:"facebook,omitempty" bson:"facebook,omitempty"`
	Google   UserOAuthInfos `json:"google,omitempty" bson:"google,omitempty"`
	Twitter  UserOAuthInfos `json:"twitter,omitempty" bson:"twitter,omitempty"`
}

// UserOAuthInfos : informations of a user from a oauth service.
type UserOAuthInfos struct {
	UserID    string    `json:"user_id" bson:"userId"`
	Token     string    `json:"token" bson:"token"`
	FirstName string    `json:"first_name" bson:"firstName,omitempty"`
	LastName  string    `json:"last_name" bson:"lastName,omitempty"`
	Email     string    `json:"email" bson:"email,omitempty"`
	CreatedAt time.Time `json:"created_at" bson:"createdAt,omitempty"`
	ExpiresAt time.Time `json:"expires_at" bson:"expiresAt,omitempty"`
}

// UserNotificationsStats : informations about the notifications received by the user.
type UserNotificationsStats struct {
	Total  string `json:"total" bson:"total"`
	Unread int    `json:"unread" bson:"unread"`
}

// NewUser : returns a new 'User' struct.
func NewUser() *User {
	return &User{
		IsAnonymous: false,
		IsAdmin:     false,
	}
}

// UserWithID : returns a struct 'user' with a ID.
func UserWithID(userID string) *User {
	user := NewUser()
	user.ID = bson.ObjectIdHex(userID)

	return user
}

// NewUserFromDisabledUser : returns a new 'User' struct.
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

// IsSaved : returns true if the 'user' document if saved in the collection.
func (user *User) IsSaved() bool {
	return user.ID.Valid()
}

// Save : inserts the 'user' document in the collection, returns an error if needed.
func (user *User) Save(db *mgo.Database) (string, error) {
	return document.SaveDocument(user, UserCollection, db)
}

// Update : updates the 'user' document, returns an error if needed.
func (user *User) Update(update interface{}, db *mgo.Database) error {
	return document.UpdateDocument(user.ID, UserCollection, update, db)
}

// Delete : deletes the 'user' document from the collection, returns an error if needed.
func (user *User) Delete(db *mgo.Database) error {
	return document.DeleteDocument(user.ID, UserCollection, db)
	// TODO: tasks to delete all the documents linked to the user.
}

// GetByID : initializes the 'User' struct from the 'user' document with the ID.
func (user *User) GetByID(id string, db *mgo.Database) error {
	return document.GetDocumentByID(user, UserCollection, id, db)
}

// GetID : returns the 'User' ID.
func (user *User) GetID() string {
	return user.ID.Hex()
}

// SetID : initializes the 'User' ID.
func (user *User) SetID(id string) {
	user.ID = bson.ObjectIdHex(id)
}

// SetAdmin : !Only for testing
func (user *User) SetAdmin(admin bool, db *mgo.Database) error {
	if config.Main == nil || config.Main.IsTestEnvironment() {
		return user.Update(bson.M{"$set": bson.M{"isAdmin": admin}}, db)
	}

	return nil
}
