package account

import (
	"errors"
	"fmt"
	"time"

	"github.com/cygy/ginamite/common/ctxnotifications"

	"github.com/cygy/ginamite/common/ctxlocalization"

	"github.com/cygy/ginamite/common/ctxqueue"

	ctxAuthToken "github.com/cygy/ginamite/api/context/authtoken"
	ctxLocale "github.com/cygy/ginamite/api/context/locale"
	ctxUser "github.com/cygy/ginamite/api/context/user"
	"github.com/cygy/ginamite/api/response"
	"github.com/cygy/ginamite/common/authentication"
	"github.com/cygy/ginamite/common/config/environment"
	"github.com/cygy/ginamite/common/ctxconfig"
	commonErrors "github.com/cygy/ginamite/common/errors"
	"github.com/cygy/ginamite/common/notifications"
	"github.com/cygy/ginamite/common/queue"
	"github.com/cygy/ginamite/common/random"
	req "github.com/cygy/ginamite/common/request"

	"github.com/gin-gonic/gin"
)

func deleteAuthenticationToken(c *gin.Context, checkKey bool) {
	locale := ctxLocale.Get(c)
	t := ctxlocalization.Get(c).Translate(locale)

	// Closure to send an invalid request error.
	sendError := func() {
		if checkKey {
			response.InvalidRequestParameterWithDetails(c,
				t("error.auth_token.delete.message"),
				t("error.auth_token.delete.reason"),
				t("error.auth_token.delete.recovery"),
			)
		} else {
			response.NotFound(c)
		}
	}

	if FunctionsOf(c).GetTokenDetailsByID == nil {
		c.Error(errors.New("undefined function 'GetTokenDetailsByID'"))
		response.InternalServerError(c)
		return
	}

	if FunctionsOf(c).DeleteTokenByID == nil {
		c.Error(errors.New("undefined function 'DeleteTokenByID'"))
		sendError()
		return
	}

	// ID and key are mandatory.
	tokenKey := c.Param("key")
	if len(tokenKey) == 0 && checkKey {
		sendError()
		return
	}

	tokenID := c.Param("id")

	// Get the requested token.
	token, err := FunctionsOf(c).GetTokenDetailsByID(c, tokenID)
	if err != nil {
		if commonErrors.IsNotFound(err) {
			sendError()
		} else {
			response.InternalServerError(c)
		}
		return
	}

	// Verify the key or the user and delete the token.
	userID := ctxUser.GetID(c)
	if (!checkKey && userID != token.UserID) || (checkKey && tokenKey != token.PrivateKey) {
		sendError()
		return
	}

	if err := FunctionsOf(c).DeleteTokenByID(c, tokenID); err != nil {
		response.InternalServerError(c)
		return
	}

	ctxqueue.Get(c).InvalidAuthToken(tokenID, token.ExpirationDate)

	// Return a no content error.
	response.OkWithStatus(c, t("status.login.cancelled"))
}

func createAuthenticationToken(c *gin.Context, user authentication.User, IPAddress, source string, device authentication.Device, method string, sendResponse, afterValidatingAccount bool) (string, error) {
	if FunctionsOf(c).SaveAuthenticationToken == nil {
		err := errors.New("undefined function 'SaveAuthenticationToken'")
		canNotLoginError(c, err)
		return "", err
	}

	if FunctionsOf(c).DeleteTokenByID == nil {
		err := errors.New("undefined function 'DeleteTokenByID'")
		canNotLoginError(c, err)
		return "", err
	}

	// Create the details of the authentication to save.
	device.NotificationSettings.Notifications = ctxnotifications.Get(c).Defaults()
	device.SetUpNameAndDetails()

	authDetails := authentication.Details{
		UserID:         user.ID,
		Device:         device,
		Source:         source,
		Method:         method,
		IPAddress:      IPAddress,
		Key:            random.String(128),
		CreationDate:   time.Now(),
		ExpirationDate: time.Now().Add(time.Duration(ctxconfig.Get(c).JWT.TTL) * time.Second),
	}

	tokenID, knownIPAddress, err := FunctionsOf(c).SaveAuthenticationToken(c, authDetails)
	if err != nil {
		canNotLoginError(c, err)
		return "", err
	}

	// Tell the workers to retrieve the IP location info.
	if !knownIPAddress {
		ctxqueue.Get(c).GetIPLocation(queue.TaskIPLocation{
			IPAddress: IPAddress,
			TokenID:   tokenID,
		})
	}

	// Sign and get the complete encoded token as a string.
	signedToken, err := ctxAuthToken.HooksOf(c).GetSignedToken(c, ctxconfig.Get(c).JWT.SigningMethod, ctxconfig.Get(c).JWT.Secret, tokenID)
	if err != nil {
		if deleteErr := FunctionsOf(c).DeleteTokenByID(c, tokenID); deleteErr != nil {
			c.Error(deleteErr)
		}
		canNotLoginError(c, err)
		return "", err
	}

	var responseContent response.H
	if ctxconfig.Get(c).Environment == environment.Test {
		responseContent = response.H{
			"auth_token": signedToken,
			"id":         tokenID,
			"key":        authDetails.Key,
		}
	} else {
		responseContent = response.H{"auth_token": signedToken}
	}

	if !afterValidatingAccount {
		ctxqueue.Get(c).NotifyUser(notifications.TypeNewLogin, queue.UserNotificationNewLogin{
			EmailAddress:   user.EmailAddress,
			UserID:         user.ID,
			Username:       user.Username,
			UnsubscribeKey: notifications.GenerateUnsubscribeKey(user.ID, user.PrivateKey, notifications.TypeNewLogin),
			Source:         source,
			Device:         fmt.Sprintf("%s (%s)", authDetails.Device.Name, authDetails.Device.Details),
			IPAddress:      req.GetRealIPAddress(c),
			TokenID:        tokenID,
			TokenKey:       authDetails.Key,
			Locale:         user.Locale,
		})
	}

	if sendResponse {
		if afterValidatingAccount {
			response.Ok(c, responseContent)
		} else {
			response.Created(c, responseContent)
		}
	}

	return signedToken, nil
}
