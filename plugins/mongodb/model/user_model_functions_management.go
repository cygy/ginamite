package model

import (
	"fmt"
	"strings"

	"github.com/cygy/ginamite/common/authentication"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

// UpdateAllInformations : updates the infos of an account.
func (user *User) UpdateAllInformations(firstName, lastName, emailAddress string, validEmailAddress bool, db *mgo.Database) error {
	return user.Update(bson.M{"$set": bson.M{"privateInfos.firstName": firstName, "privateInfos.lastName": lastName, "privateInfos.email": emailAddress, "privateInfos.isEmailValid": validEmailAddress}}, db)
}

// UpdateInfos : updates the infos of an account.
func (user *User) UpdateInfos(firstName, lastName string, db *mgo.Database) error {
	return user.Update(bson.M{"$set": bson.M{"privateInfos.firstName": firstName, "privateInfos.lastName": lastName}}, db)
}

// UpdateSettings : updates the settings of an account.
func (user *User) UpdateSettings(locale, timezone string, db *mgo.Database) error {
	return user.Update(bson.M{"$set": bson.M{"settings.locale": strings.ToLower(locale), "settings.timezone": timezone}}, db)
}

// UpdatePassword : updates the password of a user.
func (user *User) UpdatePassword(newPassword string, alreadyEncrypted bool, db *mgo.Database) error {
	var encryptedPassword string

	if !alreadyEncrypted {
		var err error
		encryptedPassword, err = authentication.EncryptPassword(newPassword)
		if err != nil {
			return err
		}
	} else {
		encryptedPassword = newPassword
	}

	return user.Update(bson.M{"$set": bson.M{"identificationMethods.password": encryptedPassword}}, db)
}

// UpdateEmailAddress : updates the email address of a user.
func (user *User) UpdateEmailAddress(emailAddress string, db *mgo.Database) error {
	return user.Update(bson.M{"$set": bson.M{"privateInfos.email": emailAddress, "privateInfos.isEmailValid": false}}, db)
}

// SetEmailAddressValid : updates the email address of a user.
func (user *User) SetEmailAddressValid(db *mgo.Database) error {
	return user.Update(bson.M{"$set": bson.M{"privateInfos.isEmailValid": true}}, db)
}

// UpdateImage : updates the image of a user.
func (user *User) UpdateImage(db *mgo.Database) error {
	return user.Update(bson.M{"$set": bson.M{
		"publicInfos.image.full.size1x.path":  user.PublicInfos.Image.Full.Size1x.Path,
		"publicInfos.image.full.size1x.url":   user.PublicInfos.Image.Full.Size1x.URL,
		"publicInfos.image.full.size2x.path":  user.PublicInfos.Image.Full.Size2x.Path,
		"publicInfos.image.full.size2x.url":   user.PublicInfos.Image.Full.Size2x.URL,
		"publicInfos.image.small.size1x.path": user.PublicInfos.Image.Small.Size1x.Path,
		"publicInfos.image.small.size1x.url":  user.PublicInfos.Image.Small.Size1x.URL,
		"publicInfos.image.small.size2x.path": user.PublicInfos.Image.Small.Size2x.Path,
		"publicInfos.image.small.size2x.url":  user.PublicInfos.Image.Small.Size2x.URL,
		"publicInfos.image.thumb.size1x.path": user.PublicInfos.Image.Thumb.Size1x.Path,
		"publicInfos.image.thumb.size1x.url":  user.PublicInfos.Image.Thumb.Size1x.URL,
		"publicInfos.image.thumb.size2x.path": user.PublicInfos.Image.Thumb.Size2x.Path,
		"publicInfos.image.thumb.size2x.url":  user.PublicInfos.Image.Thumb.Size2x.URL,
	}}, db)
}

// UpdatePublicProfile : updates the public profile of a user.
func (user *User) UpdatePublicProfile(locale, description string, db *mgo.Database) error {
	return user.Update(bson.M{"$set": bson.M{fmt.Sprintf("publicInfos.description.%s", strings.ToLower(locale)): description}}, db)
}

// UpdateSocialNetworks : updates the social network profiles of a user.
func (user *User) UpdateSocialNetworks(facebook, twitter, instagram string, db *mgo.Database) error {
	return user.Update(bson.M{"$set": bson.M{
		"publicInfos.facebook":  facebook,
		"publicInfos.twitter":   twitter,
		"publicInfos.instagram": instagram,
	}}, db)
}
