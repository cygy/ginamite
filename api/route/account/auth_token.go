package account

import (
	"errors"
	"strings"
	"time"

	ctxAuthToken "github.com/cygy/ginamite/api/context/authtoken"
	ctxLocale "github.com/cygy/ginamite/api/context/locale"
	ctxUser "github.com/cygy/ginamite/api/context/user"
	"github.com/cygy/ginamite/api/request"
	"github.com/cygy/ginamite/api/response"
	"github.com/cygy/ginamite/common/authentication"
	"github.com/cygy/ginamite/common/ctxconfig"
	commonErrors "github.com/cygy/ginamite/common/errors"
	authError "github.com/cygy/ginamite/common/errors/auth"
	"github.com/cygy/ginamite/common/localization"
	"github.com/cygy/ginamite/common/notifications"
	"github.com/cygy/ginamite/common/queue"
	req "github.com/cygy/ginamite/common/request"
	res "github.com/cygy/ginamite/common/response"

	"github.com/gin-gonic/gin"
)

// LoginWithPassword creates a token to log in an account.
func LoginWithPassword(c *gin.Context) {
	if FunctionsOf(c).GetUserDetailsAndEncryptedPasswordByIdentifierFromEnabledUser == nil {
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

	locale := ctxLocale.Get(c)
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

	user, encryptedPassword, err := FunctionsOf(c).GetUserDetailsAndEncryptedPasswordByIdentifierFromEnabledUser(c, identifier)
	if err != nil || !authentication.ComparePassword(password, encryptedPassword) {
		canNotLoginError(c, err)
		return
	}

	IPAddress := req.GetRealIPAddress(c)
	// The errors are already reported to the client.
	if _, err := createAuthenticationToken(c, user, IPAddress, source, device, authentication.MethodPassword, true, false); err != nil {
		return
	}
}

// LoginWithFacebook creates a token to log in an account.
func LoginWithFacebook(c *gin.Context) {
	loginWithThirdParty(c, newTokenValidator(c, authentication.MethodFacebook))
}

// LoginWithGoogle creates a token to log in an account.
func LoginWithGoogle(c *gin.Context) {
	loginWithThirdParty(c, newTokenValidator(c, authentication.MethodGoogle))
}

// Logout deletes the token and log out.
func Logout(c *gin.Context) {
	if FunctionsOf(c).DeleteTokenByID == nil {
		c.Error(errors.New("undefined function 'DeleteTokenByID'"))
		response.InternalServerError(c)
		return
	}

	tokenID := ctxAuthToken.GetID(c)

	expirationDate := time.Now().Add(time.Duration(ctxconfig.Get(c).JWT.TTL) * time.Second)
	if FunctionsOf(c).GetTokenExpirationDate != nil {
		if date, err := FunctionsOf(c).GetTokenExpirationDate(c, tokenID); err == nil {
			expirationDate = date
		}
	}

	if err := FunctionsOf(c).DeleteTokenByID(c, tokenID); err != nil {
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

// GetAuthTokens gets the tokens of an account.
func GetAuthTokens(c *gin.Context) {
	if FunctionsOf(c).GetOwnedTokens == nil {
		c.Error(errors.New("undefined function 'GetOwnedTokens'"))
		response.InternalServerError(c)
		return
	}

	userID := ctxUser.GetID(c)

	tokens, err := FunctionsOf(c).GetOwnedTokens(c, userID)
	if err != nil {
		response.SendError(c, err)
		return
	}

	response.Ok(c, response.H{
		"tokens": tokens,
	})
}

// DeleteAuthToken revokes a token from an account.
func DeleteAuthToken(c *gin.Context) {
	deleteAuthenticationToken(c, false)
}

// DeleteAuthTokenWithKey revokes a token from an account with a key.
func DeleteAuthTokenWithKey(c *gin.Context) {
	deleteAuthenticationToken(c, true)
}

// GetAuthToken gets the token properties of an account.
func GetAuthToken(c *gin.Context) {
	if FunctionsOf(c).GetOwnedTokenByID == nil {
		c.Error(errors.New("undefined function 'GetOwnedTokenByID'"))
		response.InternalServerError(c)
		return
	}

	tokenID := c.Param("id")

	token, tokenOwnerID, err := FunctionsOf(c).GetOwnedTokenByID(c, tokenID)
	if err != nil {
		response.SendError(c, err)
		return
	}

	userID := ctxUser.GetID(c)
	if tokenOwnerID != userID {
		response.NotFound(c)
		return
	}

	response.Ok(c, response.H{
		"token": token,
	})
}

// UpdateAuthToken updates a token of an account.
func UpdateAuthToken(c *gin.Context) {
	if FunctionsOf(c).GetTokenDetailsByID == nil {
		c.Error(errors.New("undefined function 'GetTokenDetailsByID'"))
		response.InternalServerError(c)
		return
	}

	if FunctionsOf(c).UpdateTokenByID == nil {
		c.Error(errors.New("undefined function 'UpdateTokenByID'"))
		response.InternalServerError(c)
		return
	}

	locale := ctxLocale.Get(c)
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

	token, err := FunctionsOf(c).GetTokenDetailsByID(c, tokenID)
	if err != nil {
		response.SendError(c, err)
		return
	}

	userID := ctxUser.GetID(c)
	if token.UserID != userID {
		response.NotFound(c)
		return
	}

	if err := FunctionsOf(c).UpdateTokenByID(c, tokenID, name, jsonBody.EnableNotification, subscribedNotifications); err != nil {
		response.InternalServerError(c)
		return
	}

	response.OkWithStatus(c, t("status.auth_token.updated"))
}

// RefreshAuthToken refreshes a auth token.
func RefreshAuthToken(c *gin.Context) {
	tokenID := ctxAuthToken.GetID(c)

	signedToken, err := ctxAuthToken.HooksOf(c).RefreshAuthToken(c, ctxconfig.Get(c).JWT.Secret, ctxconfig.Get(c).JWT.SigningMethod, tokenID, time.Duration(ctxconfig.Get(c).JWT.TTL))
	if err != nil {
		locale := ctxLocale.Get(c)
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
