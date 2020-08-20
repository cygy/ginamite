package functions

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	apiContext "github.com/cygy/ginamite/api/context"
	"github.com/cygy/ginamite/api/image"
	"github.com/cygy/ginamite/api/request"
	"github.com/cygy/ginamite/api/response"
	"github.com/cygy/ginamite/common/config"
	"github.com/cygy/ginamite/common/localization"
	"github.com/cygy/ginamite/common/mongo/session"
	"github.com/cygy/ginamite/common/queue"
	"github.com/cygy/ginamite/plugins/mongodb/api/context"
	"github.com/cygy/ginamite/plugins/mongodb/model"
	pluginQueue "github.com/cygy/ginamite/plugins/mongodb/queue"

	"github.com/gin-gonic/gin"
)

// GetOwnedAccountDetails : returns the details of the user's account.
func GetOwnedAccountDetails(c *gin.Context, userID string) interface{} {
	user := context.GetFullUser(c)
	if user == nil {
		return nil
	}

	return model.NewOwnUser(user)
}

// UpdatedAccountInfos : updates the details of an user's account.
func UpdatedAccountInfos(c *gin.Context, userID, firstName, lastName string) error {
	mongoSession := session.Get(c)
	user := context.GetUser(c)
	return user.UpdateInfos(firstName, lastName, mongoSession)
}

// UpdatedAccountSettings : updates the settings of an user's account.
func UpdatedAccountSettings(c *gin.Context, userID, locale, timezone string) error {
	mongoSession := session.Get(c)
	user := context.GetUser(c)
	return user.UpdateSettings(locale, timezone, mongoSession)
}

// UpdateAccountPublicProfile : updates the public profile of an account.
func UpdateAccountPublicProfile(c *gin.Context) {
	errorMessageKey := "error.account.public.profile.update.message"

	var jsonBody struct {
		Locale      string `json:"locale"`
		Description string `json:"description"`
	}
	request.DecodeBody(c, &jsonBody)

	locale := apiContext.GetLocale(c)
	t := localization.Translate(locale)

	// The locale must be present and supported.
	usedLocale := jsonBody.Locale
	if !config.Main.SupportsLocale(usedLocale) {
		response.UnsupportedParameterLocale(c, locale, usedLocale, "locale", config.Main.SupportedLocales, errorMessageKey)
		return
	}

	// Description must not be too long.
	description := jsonBody.Description
	if len(description) > int(model.UserMaxDescriptionLength) {
		response.InvalidParameterValue(c, "description",
			t(errorMessageKey),
			t("error.account.public.profile.description.too_long.reason", localization.H{"Count": model.UserMaxDescriptionLength}),
			t("error.account.public.profile.description.too_long.recovery", localization.H{"Count": model.UserMaxDescriptionLength}),
		)
		return
	}

	user := context.GetUser(c)
	mongoSession := session.Get(c)

	if err := user.UpdatePublicProfile(usedLocale, description, mongoSession); err != nil {
		response.InternalServerError(c)
		return
	}

	queue.CreateTask(pluginQueue.MessageTaskUpdateUserPublicProfile, pluginQueue.TaskUpdateUserPublicProfile{
		UserID: user.ID.Hex(),
	})

	response.OkWithStatus(c, t("status.account.public.profile.updated"))
}

// UpdateAccountPublicImage : updates the image of the public profile.
func UpdateAccountPublicImage(c *gin.Context) {
	errorMessageKey := "error.account.public.image.update.message"
	logContext := "account image"

	// Get and check the uploaded file.
	file := request.GetUploadedFile(c, "file", model.UserMaxImageSize, true, logContext, errorMessageKey)
	if file == nil {
		return
	}

	userID := apiContext.GetUserID(c)

	// Save the uploaded file to a temp file.
	rawFilename := fmt.Sprintf("%s_%d", userID, time.Now().Unix())
	hasher := md5.New()
	hasher.Write([]byte(rawFilename))
	rawFilename = hex.EncodeToString(hasher.Sum(nil))

	filename := fmt.Sprintf("%s%s", rawFilename, strings.ToLower(filepath.Ext(file.Filename)))
	tmpFilePath := filepath.Clean(os.TempDir() + "/" + filename)
	if !request.SaveUploadedFile(c, file, tmpFilePath, logContext, errorMessageKey) {
		return
	}
	defer os.Remove(tmpFilePath)

	locale := apiContext.GetLocale(c)

	// Copy the different versions of the image.
	destinationDirectory := model.UserImagePath
	versions, err := image.SaveFill(tmpFilePath, destinationDirectory,
		int(model.UserImageWidth), int(model.UserImageHeight),
		model.UserImageSmallVersion, model.UserImageThumbVersion)
	if err != nil {
		response.UnableToSaveFile(c, locale, errorMessageKey)
		return
	}

	user := context.GetFullUser(c)

	// The old images to delete.
	oldImages := []string{
		user.PublicInfos.Image.Full.Size1x.Path,
		user.PublicInfos.Image.Full.Size2x.Path,
		user.PublicInfos.Image.Small.Size1x.Path,
		user.PublicInfos.Image.Small.Size2x.Path,
		user.PublicInfos.Image.Thumb.Size1x.Path,
		user.PublicInfos.Image.Thumb.Size2x.Path,
	}

	// Set up the new images.
	user.PublicInfos.Image.Full.Size1x.Path = versions.Full.Size1x.RelativePath
	user.PublicInfos.Image.Full.Size1x.URL = config.Main.Hosts.Static + versions.Full.Size1x.AbsolutePath
	user.PublicInfos.Image.Full.Size2x.Path = versions.Full.Size2x.RelativePath
	user.PublicInfos.Image.Full.Size2x.URL = config.Main.Hosts.Static + versions.Full.Size2x.AbsolutePath
	user.PublicInfos.Image.Small.Size1x.Path = versions.Small.Size1x.RelativePath
	user.PublicInfos.Image.Small.Size1x.URL = config.Main.Hosts.Static + versions.Small.Size1x.AbsolutePath
	user.PublicInfos.Image.Small.Size2x.Path = versions.Small.Size2x.RelativePath
	user.PublicInfos.Image.Small.Size2x.URL = config.Main.Hosts.Static + versions.Small.Size2x.AbsolutePath
	user.PublicInfos.Image.Thumb.Size1x.Path = versions.Thumb.Size1x.RelativePath
	user.PublicInfos.Image.Thumb.Size1x.URL = config.Main.Hosts.Static + versions.Thumb.Size1x.AbsolutePath
	user.PublicInfos.Image.Thumb.Size2x.Path = versions.Thumb.Size2x.RelativePath
	user.PublicInfos.Image.Thumb.Size2x.URL = config.Main.Hosts.Static + versions.Thumb.Size2x.AbsolutePath

	mongoSession := session.Get(c)

	if err := user.UpdateImage(mongoSession); err != nil {
		image.Delete([]string{
			versions.Full.Size1x.RelativePath,
			versions.Full.Size2x.RelativePath,
			versions.Small.Size1x.RelativePath,
			versions.Small.Size2x.RelativePath,
			versions.Thumb.Size1x.RelativePath,
			versions.Thumb.Size2x.RelativePath,
		}, destinationDirectory)
		response.UnableToUpdateDatabase(c, locale, errorMessageKey)
		return
	}

	// Delete the old images.
	image.Delete(oldImages, destinationDirectory)

	// The auth token is updated too with the updated image of the user.
	apiContext.RefreshAuthToken(c)

	queue.CreateTask(pluginQueue.MessageTaskUpdateUserPublicImage, pluginQueue.TaskUpdateUserPublicImage{
		UserID: user.ID.Hex(),
	})

	response.Ok(c, response.H{
		"status": localization.Translate(locale)("status.account.public.image.updated"),
		"image":  model.NewOwnImage(user.PublicInfos.Image),
	})
}

// DeleteAccountPublicImage : deletes the image of the public profile.
func DeleteAccountPublicImage(c *gin.Context) {
	locale := apiContext.GetLocale(c)
	user := context.GetFullUser(c)
	mongoSession := session.Get(c)

	// The images to delete.
	oldImages := []string{
		user.PublicInfos.Image.Full.Size1x.Path,
		user.PublicInfos.Image.Full.Size2x.Path,
		user.PublicInfos.Image.Small.Size1x.Path,
		user.PublicInfos.Image.Small.Size2x.Path,
		user.PublicInfos.Image.Thumb.Size1x.Path,
		user.PublicInfos.Image.Thumb.Size2x.Path,
	}

	// Remove the images.
	user.PublicInfos.Image = model.Image{}
	if err := user.UpdateImage(mongoSession); err != nil {
		response.UnableToUpdateDatabase(c, locale, "error.account.public.image.delete.message")
		return
	}

	// Delete the files.
	image.Delete(oldImages, model.UserImagePath)

	// The auth token is updated too with the updated image of the user.
	apiContext.RefreshAuthToken(c)

	queue.CreateTask(pluginQueue.MessageTaskUpdateUserPublicImage, pluginQueue.TaskUpdateUserPublicImage{
		UserID: user.ID.Hex(),
	})

	response.OkWithStatus(c, localization.Translate(locale)("status.account.public.image.deleted"))
}
