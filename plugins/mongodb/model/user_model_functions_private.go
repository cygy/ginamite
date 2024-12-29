package model

import (
	"fmt"
	"strings"
	"time"

	"github.com/cygy/ginamite/common/authentication"
	"github.com/cygy/ginamite/common/config"
	"github.com/cygy/ginamite/common/notifications"
	"github.com/cygy/ginamite/common/timezone"
	"github.com/gosimple/slug"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

func (user *User) saveThirdPartyInfos(source, token, userID, firstName, lastName, email string, createdAt, expiresAt time.Time, db *mgo.Database) error {
	infos := UserOAuthInfos{
		UserID:    userID,
		Token:     token,
		FirstName: firstName,
		LastName:  lastName,
		Email:     email,
		CreatedAt: createdAt,
		ExpiresAt: expiresAt,
	}

	updatedFields := bson.M{fmt.Sprintf("identificationMethods.%s", source): infos}

	// Update the first name and the last name if there are provided.
	if (len(user.PrivateInfos.FirstName) == 0 || len(user.PrivateInfos.LastName) == 0) && (len(infos.FirstName) > 0 && len(infos.LastName) > 0) {
		updatedFields["privateInfos.firstName"] = firstName
		updatedFields["privateInfos.lastName"] = lastName
	}

	// If the email address is validated by a third party, valid it here too.
	if !user.PrivateInfos.IsEmailValid && (len(email) > 0) && strings.EqualFold(email, user.PrivateInfos.Email) {
		updatedFields["privateInfos.isEmailValid"] = true
	}

	return user.Update(bson.M{"$set": updatedFields}, db)
}

func (user *User) initializeFromRegistration(username, emailAddress, termsVersion, locale, registrationCode, privateKey, method, source, ip string, device authentication.Device) {
	parts := strings.SplitN(locale, "-", 2)
	if len(parts) > 1 {
		country := parts[1]
		timezones := timezone.Main.GetTimezones(country)
		if len(timezones) > 0 {
			user.Settings.Timezone = timezones[0]
		} else if config.Main != nil {
			user.Settings.Timezone = config.Main.DefaultTimezone
		}
	}

	userDevice := Device{}
	userDevice.FromDevice(device)

	user.Username = username
	user.Slug = slug.Make(username)
	user.Abilities = DefaultUserAbilities
	user.Notifications = notifications.DefaultNotifications()
	user.PrivateInfos.Email = emailAddress
	user.PrivateInfos.IsEmailValid = false
	user.RegistrationInfos.Date = time.Now()
	user.RegistrationInfos.Code = registrationCode
	user.RegistrationInfos.Source = source
	user.RegistrationInfos.Device = userDevice
	user.RegistrationInfos.Method = method
	user.RegistrationInfos.IPAddress = ip
	user.Settings.Locale = locale
	user.Verified = false
	user.VersionOfTermsAccepted = termsVersion
	user.PrivateKey = privateKey
}
