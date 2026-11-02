package model

import (
	"context"
	"fmt"
	"strings"

	"github.com/cygy/ginamite/common/authentication"
	"github.com/cygy/ginamite/plugins/mongodb/mongo/bson"
	"github.com/cygy/ginamite/plugins/mongodb/mongo/store"
)

// UpdateAllInformations updates the infos of an account.
func (user *User) UpdateAllInformations(ctx context.Context, firstName, lastName, emailAddress string, validEmailAddress bool, db store.Database) error {
	return user.Update(ctx, bson.M{"$set": bson.M{"privateInfos.firstName": firstName, "privateInfos.lastName": lastName, "privateInfos.email": emailAddress, "privateInfos.isEmailValid": validEmailAddress}}, db)
}

// UpdateInfos updates the infos of an account.
func (user *User) UpdateInfos(ctx context.Context, firstName, lastName string, db store.Database) error {
	return user.Update(ctx, bson.M{"$set": bson.M{"privateInfos.firstName": firstName, "privateInfos.lastName": lastName}}, db)
}

// UpdateSettings updates the settings of an account.
func (user *User) UpdateSettings(ctx context.Context, locale, timezone string, db store.Database) error {
	return user.Update(ctx, bson.M{"$set": bson.M{"settings.locale": strings.ToLower(locale), "settings.timezone": timezone}}, db)
}

// UpdatePassword updates the password of a user.
func (user *User) UpdatePassword(ctx context.Context, newPassword string, alreadyEncrypted bool, db store.Database) error {
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

	return user.Update(ctx, bson.M{"$set": bson.M{"identificationMethods.password": encryptedPassword}}, db)
}

// UpdateEmailAddress updates the email address of a user.
func (user *User) UpdateEmailAddress(ctx context.Context, emailAddress string, db store.Database) error {
	return user.Update(ctx, bson.M{"$set": bson.M{"privateInfos.email": emailAddress, "privateInfos.isEmailValid": false}}, db)
}

// SetEmailAddressValid marks the email address of the user as valid.
func (user *User) SetEmailAddressValid(ctx context.Context, db store.Database) error {
	return user.Update(ctx, bson.M{"$set": bson.M{"privateInfos.isEmailValid": true}}, db)
}

// UpdateImage updates the image of a user.
func (user *User) UpdateImage(ctx context.Context, db store.Database) error {
	return user.Update(ctx, bson.M{"$set": bson.M{
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

// UpdatePublicProfile updates the public profile of a user.
func (user *User) UpdatePublicProfile(ctx context.Context, locale, description string, db store.Database) error {
	return user.Update(ctx, bson.M{"$set": bson.M{fmt.Sprintf("publicInfos.description.%s", strings.ToLower(locale)): description}}, db)
}

// UpdateSocialNetworks updates the social network profiles of a user.
func (user *User) UpdateSocialNetworks(ctx context.Context, facebook, twitter, instagram string, db store.Database) error {
	return user.Update(ctx, bson.M{"$set": bson.M{
		"publicInfos.facebook":  facebook,
		"publicInfos.twitter":   twitter,
		"publicInfos.instagram": instagram,
	}}, db)
}

// AcceptVersionOfTerms updates the version of the terms accepted.
func (user *User) AcceptVersionOfTerms(ctx context.Context, version string, db store.Database) error {
	return user.Update(ctx, bson.M{"$set": bson.M{"versionOfTermsAccepted": version}}, db)
}
