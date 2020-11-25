package account

import (
	"errors"
	"strings"
	"time"

	"github.com/cygy/ginamite/api/context"
	"github.com/cygy/ginamite/api/request"
	"github.com/cygy/ginamite/api/response"
	"github.com/cygy/ginamite/api/validator"
	"github.com/cygy/ginamite/common/authentication"
	"github.com/cygy/ginamite/common/config"
	commonErrors "github.com/cygy/ginamite/common/errors"
	authError "github.com/cygy/ginamite/common/errors/auth"
	"github.com/cygy/ginamite/common/localization"
	"github.com/cygy/ginamite/common/notifications"
	"github.com/cygy/ginamite/common/queue"
	req "github.com/cygy/ginamite/common/request"
	res "github.com/cygy/ginamite/common/response"

	"github.com/gin-gonic/gin"
)

// LoginWithPassword : creates a token to log in an account.
func LoginWithPassword(c *gin.Context) {
	if GetUserDetailsAndEncryptedPasswordByIdentifierFromEnabledUser == nil {
		c.Error(errors.New("undefined function 'GetUserDetailsAndEncryptedPasswordByIdentifierFromEnabledUser'"))
		response.InternalServerError(c)
		return
	}

	var jsonBody struct {
		Identifier string                `json:"identifier"` // can be the username or the email address
		Password   string                `json:"password"`
		Source     string                `json:"source"`
		Device     authentication.Device `json:"device"`
	}
	request.DecodeBody(c, &jsonBody)

	locale := context.GetLocale(c)
	t := localization.Translate(locale)

	// Identifier is mandatory.
	identifier := jsonBody.Identifier
	if len(identifier) == 0 {
		message := t("error.login.unable.message")
		reason := t("error.login.identifier.not_found.reason")
		recovery := t("error.login.identifier.not_found.recovery")

		response.NotFoundParameterValue(c, "identifier", message, reason, recovery)
		return
	}

	// Password is mandatory.
	password := jsonBody.Password
	if len(password) == 0 {
		message := t("error.login.unable.message")
		reason := t("error.login.password.not_found.reason")
		recovery := t("error.login.password.not_found.recovery")

		response.NotFoundParameterValue(c, "password", message, reason, recovery)
		return
	}

	// Source is mandatory.
	source := jsonBody.Source
	if len(source) == 0 {
		message := t("error.login.unable.message")
		reason := t("error.login.source.not_found.reason")
		recovery := t("error.login.source.not_found.recovery")

		response.NotFoundParameterValue(c, "source", message, reason, recovery)
		return
	}

	// Device is mandatory.
	device := jsonBody.Device
	if len(device.Type) == 0 || len(device.Name) == 0 {
		message := t("error.login.unable.message")
		reason := t("error.login.device.not_found.reason")
		recovery := t("error.login.device.not_found.recovery")

		response.NotFoundParameterValue(c, "device", message, reason, recovery)
		return
	}

	user, encryptedPassword, err := GetUserDetailsAndEncryptedPasswordByIdentifierFromEnabledUser(c, identifier)
	if err != nil || !authentication.ComparePassword(password, encryptedPassword) {
		canNotLoginError(c, err)
		return
	}

	IPAddress := req.GetRealIPAddress(c)
	createAuthenticationToken(c, user, IPAddress, source, device, authentication.MethodPassword, true, false)
}

// LoginWithFacebook : creates a token to log in an account.
func LoginWithFacebook(c *gin.Context) {
	v := validator.NewFacebookTokenValidator(config.Main.Facebook.AppID, config.Main.Facebook.APIVersion, config.Main.Facebook.TimeOut, config.Main.Facebook.RetryCount, config.Main.IsDebugModeEnabled())
	loginWithThirdParty(c, v)
}

// LoginWithGoogle : creates a token to log in an account.
func LoginWithGoogle(c *gin.Context) {
	v := validator.NewGoogleTokenValidator(config.Main.Google.ClientID, config.Main.Google.APIVersion, config.Main.Google.TimeOut, config.Main.Google.RetryCount, config.Main.IsDebugModeEnabled())
	loginWithThirdParty(c, v)
}

// Logout : deletes the token and log out.
func Logout(c *gin.Context) {
	if DeleteTokenByID == nil {
		c.Error(errors.New("undefined function 'DeleteTokenByID'"))
		response.InternalServerError(c)
		return
	}

	tokenID := context.GetAuthTokenID(c)

	expirationDate := time.Now().Add(time.Duration(config.Main.JWT.TTL) * time.Second)
	if GetTokenExpirationDate != nil {
		if date, err := GetTokenExpirationDate(c, tokenID); err != nil {
			expirationDate = date
		}
	}

	if err := DeleteTokenByID(c, tokenID); err != nil {
		if commonErrors.IsNotFound(err) {
			response.NoContent(c)
		} else {
			response.InternalServerError(c)
		}
		return
	}

	queue.InvalidAuthToken(tokenID, expirationDate)

	response.NoContent(c)
}

// GetAuthTokens : gets the tokens of an account.
func GetAuthTokens(c *gin.Context) {
	if GetOwnedTokens == nil {
		c.Error(errors.New("undefined function 'GetOwnedTokens'"))
		response.InternalServerError(c)
		return
	}

	userID := context.GetUserID(c)

	tokens, err := GetOwnedTokens(c, userID)
	if err != nil {
		response.SendError(c, err)
		return
	}

	response.Ok(c, response.H{
		"tokens": tokens,
	})
}

// DeleteAuthToken : revokes a token from an account.
func DeleteAuthToken(c *gin.Context) {
	deleteAuthenticationToken(c, false)
}

// DeleteAuthTokenWithKey : revokes a token from an account with a key.
func DeleteAuthTokenWithKey(c *gin.Context) {
	deleteAuthenticationToken(c, true)
}

// GetAuthToken : gets the token properties of an account.
func GetAuthToken(c *gin.Context) {
	if GetOwnedTokenByID == nil {
		c.Error(errors.New("undefined function 'GetOwnedTokenByID'"))
		response.InternalServerError(c)
		return
	}

	tokenID := c.Param("id")

	token, tokenOwnerID, err := GetOwnedTokenByID(c, tokenID)
	if err != nil {
		response.SendError(c, err)
		return
	}

	userID := context.GetUserID(c)
	if tokenOwnerID != userID {
		response.NotFound(c)
		return
	}

	response.Ok(c, response.H{
		"token": token,
	})
}

// UpdateAuthToken : updates a token of an account.
func UpdateAuthToken(c *gin.Context) {
	if GetTokenDetailsByID == nil {
		c.Error(errors.New("undefined function 'GetTokenDetailsByID'"))
		response.InternalServerError(c)
		return
	}

	if UpdateTokenByID == nil {
		c.Error(errors.New("undefined function 'UpdateTokenByID'"))
		response.InternalServerError(c)
		return
	}

	locale := context.GetLocale(c)
	t := localization.Translate(locale)

	tokenID := c.Param("id")

	var jsonBody struct {
		Name               string   `json:"name"`
		EnableNotification bool     `json:"enable_notification"`
		Notifications      []string `json:"notifications"`
	}
	request.DecodeBody(c, &jsonBody)

	// Name is mandatory.
	name := jsonBody.Name
	if len(name) == 0 {
		message := t("error.auth_token.update.unable.message")
		reason := t("error.auth_token.update.name.not_found.reason")
		recovery := t("error.auth_token.update.name.not_found.recovery")

		response.NotFoundParameterValue(c, "name", message, reason, recovery)
		return
	}

	subscribedNotifications := jsonBody.Notifications
	for _, notification := range subscribedNotifications {
		if !notifications.IsNotificationSupported(notification) {
			message := t("error.auth_token.update.unable.message")
			reason := t("error.auth_token.update.notifications.invalid.reason", localization.H{"Value": notification})
			recovery := t("error.auth_token.update.notifications.invalid.recovery", localization.H{"Value": strings.Join(notifications.SupportedNotifications(), ",")})

			response.InvalidParameterValue(c, "notifications", message, reason, recovery)
			return
		}
	}

	token, err := GetTokenDetailsByID(c, tokenID)
	if err != nil {
		response.SendError(c, err)
		return
	}

	userID := context.GetUserID(c)
	if token.UserID != userID {
		response.NotFound(c)
		return
	}

	if err := UpdateTokenByID(c, tokenID, name, jsonBody.EnableNotification, subscribedNotifications); err != nil {
		response.InternalServerError(c)
		return
	}

	response.OkWithStatus(c, t("status.auth_token.updated"))
}

// RefreshAuthToken : refreshes a auth token.
func RefreshAuthToken(c *gin.Context) {
	tokenID := context.GetAuthTokenID(c)

	signedToken, err := authentication.RefreshAuthToken(config.Main.JWT.Secret, config.Main.JWT.SigningMethod, tokenID, time.Duration(config.Main.JWT.TTL))
	if err != nil {
		locale := context.GetLocale(c)
		t := localization.Translate(locale)

		response.BadRequestWithError(c, res.Error{
			Domain:   authError.Domain,
			Code:     authError.CanNotRefresh,
			Message:  t("error.auth_token.refresh.message"),
			Reason:   t("error.auth_token.refresh.reason"),
			Recovery: t("error.auth_token.refresh.recovery"),
		})
		return
	}

	response.Ok(c, response.H{
		"auth_token": signedToken,
	})
}
