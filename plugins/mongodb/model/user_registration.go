package model

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/cygy/ginamite/api/validator"
	"github.com/cygy/ginamite/common/authentication"
	"github.com/cygy/ginamite/common/log"
	"github.com/cygy/ginamite/common/mongo/bson"
	"github.com/cygy/ginamite/common/mongo/store"
	"github.com/cygy/ginamite/common/timezone"

	"github.com/gosimple/slug"
)

// RegistrationTimezone returns the timezone of a new user: the first timezone of the country of the locale, such as
// "fr-FR", or the default timezone if there is none. The timezones, if nil, are not used. It returns an empty string if
// the locale has no country.
func RegistrationTimezone(locale string, timezones *timezone.Configuration, defaultTimezone string) string {
	parts := strings.SplitN(locale, "-", 2)
	if len(parts) < 2 {
		return ""
	}

	if timezones != nil {
		if countryTimezones := timezones.GetTimezones(parts[1]); len(countryTimezones) > 0 {
			return countryTimezones[0]
		}
	}

	return defaultTimezone
}

// RegistrationDefaults are the values given to a new user.
type RegistrationDefaults struct {
	Timezone      string   // See RegistrationTimezone.
	Notifications []string // The notifications the user is subscribed to.
}

// RegisterByEmailAddress registers a new user with an email address, with the defaults.
func (user *User) RegisterByEmailAddress(ctx context.Context, username, encryptedPassword, emailAddress, locale, termsVersion, registrationCode, privateKey, source, ip string, device authentication.Device, defaults RegistrationDefaults, db store.Database) error {
	user.initializeFromRegistration(username, emailAddress, termsVersion, locale, registrationCode, privateKey, authentication.MethodPassword, source, ip, defaults, device)
	user.IdentificationMethods.Password = encryptedPassword

	_, err := user.Save(ctx, db)
	return err
}

// ValidateRegistrationByEmailAddress validates the registration by email address.
// A wrong code is counted as an attempt; the registration is cleared when it is the UserMaxValidateRegistrationAttemps-th one.
func (user *User) ValidateRegistrationByEmailAddress(ctx context.Context, code string, db store.Database) error {
	// Wrong verification code.
	if user.RegistrationInfos.Code != code {
		if user.RegistrationInfos.CodeAttempts+1 < UserMaxValidateRegistrationAttemps {
			if err := user.Update(ctx, bson.M{"$inc": bson.M{"registrationInfos.codeAttempts": 1}}, db); err != nil {
				log.WithField("error", err).Error("unable to update")
			}
		} else {
			if err := user.Update(ctx, bson.M{"$unset": bson.M{"registrationInfos.source": "", "registrationInfos.device": "", "registrationInfos.method": "", "registrationInfos.code": "", "registrationInfos.codeAttempts": ""}}, db); err != nil {
				log.WithField("error", err).Error("unable to update")
			}
		}

		return errors.New("registration code is different")
	}

	// Verification code is OK: enable the account.
	return user.Update(ctx, bson.M{"$set": bson.M{"privateInfos.isEmailValid": true, "verified": true}, "$unset": bson.M{"registrationInfos.source": "", "registrationInfos.device": "", "registrationInfos.method": "", "registrationInfos.code": "", "registrationInfos.codeAttempts": ""}}, db)
}

// CancelRegistrationByEmailAddress cancels the registration by email address.
func (user *User) CancelRegistrationByEmailAddress(ctx context.Context, code string, db store.Database) error {
	if user.RegistrationInfos.Code != code {
		return errors.New("registration code is different")
	}

	return user.Delete(ctx, db)
}

// RegisterByThirdPartyToken registers a new user with a third-party identification token, with the defaults.
func (user *User) RegisterByThirdPartyToken(ctx context.Context, username string, tokenInfos validator.TokenInfos, tokenSource, locale, termsVersion, registrationCode, privateKey, source, ip string, device authentication.Device, defaults RegistrationDefaults, db store.Database) error {
	user.initializeFromRegistration(username, tokenInfos.Email, termsVersion, locale, registrationCode, privateKey, tokenSource, source, ip, defaults, device)

	userOAuthInfos := UserOAuthInfos{}
	userOAuthInfos.FromSocialNetworkOAuthInfos(tokenInfos)

	switch tokenSource {
	case authentication.MethodFacebook:
		user.IdentificationMethods.Facebook = userOAuthInfos
	case authentication.MethodGoogle:
		user.IdentificationMethods.Google = userOAuthInfos
	default:
		break
	}

	user.PrivateInfos.FirstName = tokenInfos.FirstName
	user.PrivateInfos.LastName = tokenInfos.LastName
	user.PrivateInfos.IsEmailValid = true
	user.Verified = true

	_, err := user.Save(ctx, db)
	return err
}

func (user *User) saveThirdPartyInfos(ctx context.Context, source, token, userID, firstName, lastName, email string, createdAt, expiresAt time.Time, db store.Database) error {
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

	return user.Update(ctx, bson.M{"$set": updatedFields}, db)
}

func (user *User) initializeFromRegistration(username, emailAddress, termsVersion, locale, registrationCode, privateKey, method, source, ip string, defaults RegistrationDefaults, device authentication.Device) {
	user.Settings.Timezone = defaults.Timezone

	userDevice := Device{}
	userDevice.FromDevice(device)

	user.Username = username
	user.Slug = slug.Make(username)
	user.Abilities = slices.Clone(DefaultUserAbilities)
	user.Notifications = defaults.Notifications
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
