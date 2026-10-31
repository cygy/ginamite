package functions

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cygy/ginamite/plugins/mongodb/api/ctxsettings"

	"github.com/cygy/ginamite/common/ctxlocalization"

	"github.com/cygy/ginamite/common/ctxqueue"

	ctxAuthToken "github.com/cygy/ginamite/api/context/authtoken"
	ctxLocale "github.com/cygy/ginamite/api/context/locale"
	ctxUser "github.com/cygy/ginamite/api/context/user"
	"github.com/cygy/ginamite/api/image"
	"github.com/cygy/ginamite/api/request"
	"github.com/cygy/ginamite/api/response"
	"github.com/cygy/ginamite/common/ctxconfig"
	"github.com/cygy/ginamite/common/log"
	"github.com/cygy/ginamite/common/mongo/session"
	"github.com/cygy/ginamite/plugins/mongodb/api/context"
	"github.com/cygy/ginamite/plugins/mongodb/model"
	pluginQueue "github.com/cygy/ginamite/plugins/mongodb/queue"
	"github.com/gin-gonic/gin"
)

// UpdateAccountPublicImage updates the image of the public profile.
func UpdateAccountPublicImage(c *gin.Context) {
	settings := ctxsettings.GetSettings(c)
	errorMessageKey := "error.account.public.image.update.message"
	logContext := "account image"

	// Get and check the uploaded file.
	file := request.GetUploadedFile(c, "file", settings.UserMaxImageSize, true, logContext, errorMessageKey)
	if file == nil {
		return
	}

	// The user is loaded first: nothing is saved if it can not be.
	user := context.GetFullUser(c)
	if user == nil {
		response.InternalServerError(c)
		return
	}

	userID := ctxUser.GetID(c)

	// Save the uploaded file to a temp file.
	rawFilename := fmt.Sprintf("%s_%d", userID, time.Now().Unix())
	hasher := md5.New()
	hasher.Write([]byte(rawFilename))
	rawFilename = hex.EncodeToString(hasher.Sum(nil))

	filename := fmt.Sprintf("%s%s", rawFilename, strings.ToLower(filepath.Ext(file.Filename)))

	// The file is saved in its own directory: gin changes the permissions of the directory of the file, and that
	// is not allowed on the shared temporary directory.
	tmpDirectory, err := os.MkdirTemp("", "account-image-")
	if err != nil {
		log.WithField("error", err).Error("unable to create the temporary directory")
		response.UnableToSaveFile(c, ctxLocale.Get(c), errorMessageKey)
		return
	}
	defer func() {
		if err := os.RemoveAll(tmpDirectory); err != nil {
			log.WithField("error", err).Error("unable to remove the temporary directory")
		}
	}()

	tmpFilePath := filepath.Join(tmpDirectory, filename)
	if !request.SaveUploadedFile(c, file, tmpFilePath, logContext, errorMessageKey) {
		return
	}

	locale := ctxLocale.Get(c)

	// Copy the different versions of the image.
	destinationDirectory := settings.UserImagePath
	versions, err := image.SaveFill(tmpFilePath, destinationDirectory,
		int(settings.UserImageWidth), int(settings.UserImageHeight),
		settings.UserImageSmallVersion, settings.UserImageThumbVersion)
	if err != nil {
		response.UnableToSaveFile(c, locale, errorMessageKey)
		return
	}

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
	user.PublicInfos.Image.Full.Size1x.URL = ctxconfig.Get(c).Hosts.Static + versions.Full.Size1x.AbsolutePath
	user.PublicInfos.Image.Full.Size2x.Path = versions.Full.Size2x.RelativePath
	user.PublicInfos.Image.Full.Size2x.URL = ctxconfig.Get(c).Hosts.Static + versions.Full.Size2x.AbsolutePath
	user.PublicInfos.Image.Small.Size1x.Path = versions.Small.Size1x.RelativePath
	user.PublicInfos.Image.Small.Size1x.URL = ctxconfig.Get(c).Hosts.Static + versions.Small.Size1x.AbsolutePath
	user.PublicInfos.Image.Small.Size2x.Path = versions.Small.Size2x.RelativePath
	user.PublicInfos.Image.Small.Size2x.URL = ctxconfig.Get(c).Hosts.Static + versions.Small.Size2x.AbsolutePath
	user.PublicInfos.Image.Thumb.Size1x.Path = versions.Thumb.Size1x.RelativePath
	user.PublicInfos.Image.Thumb.Size1x.URL = ctxconfig.Get(c).Hosts.Static + versions.Thumb.Size1x.AbsolutePath
	user.PublicInfos.Image.Thumb.Size2x.Path = versions.Thumb.Size2x.RelativePath
	user.PublicInfos.Image.Thumb.Size2x.URL = ctxconfig.Get(c).Hosts.Static + versions.Thumb.Size2x.AbsolutePath

	mongoSession := session.Get(c)

	if err := user.UpdateImage(c, mongoSession); err != nil {
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
	ctxAuthToken.Refresh(c)

	ctxqueue.Get(c).CreateTask(pluginQueue.MessageTaskUpdateUserPublicImage, pluginQueue.TaskUpdateUserPublicImage{
		UserID: user.ID.Hex(),
	})

	response.Ok(c, response.H{
		"status": ctxlocalization.Get(c).Translate(locale)("status.account.public.image.updated"),
		"image":  model.NewOwnImage(user.PublicInfos.Image),
	})
}

// DeleteAccountPublicImage deletes the image of the public profile.
func DeleteAccountPublicImage(c *gin.Context) {
	settings := ctxsettings.GetSettings(c)
	locale := ctxLocale.Get(c)
	user := context.GetFullUser(c)
	if user == nil {
		response.InternalServerError(c)
		return
	}

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
	if err := user.UpdateImage(c, mongoSession); err != nil {
		response.UnableToUpdateDatabase(c, locale, "error.account.public.image.delete.message")
		return
	}

	// Delete the files.
	image.Delete(oldImages, settings.UserImagePath)

	// The auth token is updated too with the updated image of the user.
	ctxAuthToken.Refresh(c)

	ctxqueue.Get(c).CreateTask(pluginQueue.MessageTaskUpdateUserPublicImage, pluginQueue.TaskUpdateUserPublicImage{
		UserID: user.ID.Hex(),
	})

	response.OkWithStatus(c, ctxlocalization.Get(c).Translate(locale)("status.account.public.image.deleted"))
}
