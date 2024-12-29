package model

import (
	"errors"

	"github.com/cygy/ginamite/api/validator"
	"github.com/cygy/ginamite/common/authentication"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

// RegisterByEmailAddress : register a new user with an email address.
func (user *User) RegisterByEmailAddress(username, encryptedPassword, emailAddress, locale, termsVersion, registrationCode, privateKey, source, ip string, device authentication.Device, db *mgo.Database) error {
	user.initializeFromRegistration(username, emailAddress, termsVersion, locale, registrationCode, privateKey, authentication.MethodPassword, source, ip, device)
	user.IdentificationMethods.Password = encryptedPassword

	_, err := user.Save(db)
	return err
}

// ValidateRegistrationByEmailAddress : validate the registration by email address.
func (user *User) ValidateRegistrationByEmailAddress(code string, db *mgo.Database) error {
	// Wrong verification code.
	if user.RegistrationInfos.Code != code {
		if user.RegistrationInfos.CodeAttempts <= UserMaxValidateRegistrationAttemps {
			user.Update(bson.M{"$inc": bson.M{"registrationInfos.codeAttempts": 1}}, db)
		} else {
			user.Update(bson.M{"$unset": bson.M{"registrationInfos.source": "", "registrationInfos.device": "", "registrationInfos.method": "", "registrationInfos.code": "", "registrationInfos.codeAttempts": ""}}, db)
		}

		return errors.New("registration code is different")
	}

	// Verification code is OK: enable the account.
	return user.Update(bson.M{"$set": bson.M{"privateInfos.isEmailValid": true, "verified": true}, "$unset": bson.M{"registrationInfos.source": "", "registrationInfos.device": "", "registrationInfos.method": "", "registrationInfos.code": "", "registrationInfos.codeAttempts": ""}}, db)
}

// CancelRegistrationByEmailAddress : cancel the registration by email address.
func (user *User) CancelRegistrationByEmailAddress(code string, db *mgo.Database) error {
	if user.RegistrationInfos.Code != code {
		return errors.New("registration code is different")
	}

	return user.Delete(db)
}

// RegisterByThirdPartyToken : register a new user with a third-party identification token.
func (user *User) RegisterByThirdPartyToken(username string, tokenInfos validator.TokenInfos, tokenSource, locale, termsVersion, registrationCode, privateKey, source, ip string, device authentication.Device, db *mgo.Database) error {
	user.initializeFromRegistration(username, tokenInfos.Email, termsVersion, locale, registrationCode, privateKey, tokenSource, source, ip, device)

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

	_, err := user.Save(db)
	return err
}
