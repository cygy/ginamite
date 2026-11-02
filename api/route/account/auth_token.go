package account

import (
	"errors"
	"strings"
	"time"

	"github.com/cygy/ginamite/common/ctxkey"

	ctxAuthToken "github.com/cygy/ginamite/api/context/authtoken"
	ctxLocale "github.com/cygy/ginamite/api/context/locale"
	ctxUser "github.com/cygy/ginamite/api/context/user"
	"github.com/cygy/ginamite/api/request"
	"github.com/cygy/ginamite/api/response"
	"github.com/cygy/ginamite/common/authentication"
	commonErrors "github.com/cygy/ginamite/common/errors"
	authError "github.com/cygy/ginamite/common/errors/auth"
	"github.com/cygy/ginamite/common/localization"
	"github.com/cygy/ginamite/common/log"
	res "github.com/cygy/ginamite/common/response"

	"github.com/gin-gonic/gin"
)

// LoginWithPassword creates a token to log in an account.
func LoginWithPassword(c *gin.Context) {
	if StorageOf(c).Accounts == nil {
		c.Error(errNoAccounts)
		response.InternalServerError(c)
		return
	}

	var jsonBody struct {
		Identifier string                `json:"identifier"` // can be the username or the email address
		Password   string                `json:"password"`
		Source     string                `json:"source"`
		Device     authentication.Device `json:"device"`
	}
	if !request.DecodeBody(c, &jsonBody) {
		return
	}

	locale := ctxLocale.Get(c)
	t := ctxkey.Localization.Get(c).Translate(locale)

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

	user, encryptedPassword, err := StorageOf(c).Accounts.GetUserDetailsAndEncryptedPasswordByIdentifierFromEnabledUser(c, identifier)
	if err != nil {
		// Spend the time of a comparison: a login must not be faster when the account does not exist.
		authentication.DummyCompare(password)
		canNotLoginError(c, err)
		return
	}
	if !authentication.ComparePassword(password, encryptedPassword) {
		canNotLoginError(c, nil)
		return
	}

	upgradePassword(c, user.ID, password, encryptedPassword)

	IPAddress := c.ClientIP()
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

// errNoTokens is logged when a route of the auth tokens is called without storage of the tokens.
var errNoTokens = errors.New("undefined storage of the auth tokens")

// Logout deletes the token and log out.
func Logout(c *gin.Context) {
	if StorageOf(c).Tokens == nil {
		c.Error(errNoTokens)
		response.InternalServerError(c)
		return
	}

	tokenID := ctxAuthToken.GetID(c)

	expirationDate := time.Now().Add(time.Duration(ctxkey.Config.Get(c).JWT.TTL) * time.Second)
	if date, err := StorageOf(c).Tokens.GetTokenExpirationDate(c, tokenID); err == nil {
		expirationDate = date
	}

	if err := StorageOf(c).Tokens.DeleteTokenByID(c, tokenID); err != nil {
		if commonErrors.IsNotFound(err) {
			response.NoContent(c)
		} else {
			response.InternalServerError(c)
		}
		return
	}

	ctxkey.Queue.Get(c).InvalidAuthToken(tokenID, expirationDate)

	response.NoContent(c)
}

// GetAuthTokens gets the tokens of an account.
func GetAuthTokens(c *gin.Context) {
	if StorageOf(c).Tokens == nil {
		c.Error(errNoTokens)
		response.InternalServerError(c)
		return
	}

	userID := ctxUser.GetID(c)

	tokens, err := StorageOf(c).Tokens.GetOwnedTokens(c, userID)
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
	if StorageOf(c).Tokens == nil {
		c.Error(errNoTokens)
		response.InternalServerError(c)
		return
	}

	tokenID := c.Param("id")

	token, tokenOwnerID, err := StorageOf(c).Tokens.GetOwnedTokenByID(c, tokenID)
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
	if StorageOf(c).Tokens == nil {
		c.Error(errNoTokens)
		response.InternalServerError(c)
		return
	}

	locale := ctxLocale.Get(c)
	t := ctxkey.Localization.Get(c).Translate(locale)

	tokenID := c.Param("id")

	var jsonBody struct {
		Name               string   `json:"name"`
		EnableNotification bool     `json:"enable_notification"`
		Notifications      []string `json:"notifications"`
	}
	if !request.DecodeBody(c, &jsonBody) {
		return
	}

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
		if !ctxkey.Notifications.Get(c).IsSupported(notification) {
			message := t("error.auth_token.update.unable.message")
			reason := t("error.auth_token.update.notifications.invalid.reason", localization.H{"Value": notification})
			recovery := t("error.auth_token.update.notifications.invalid.recovery", localization.H{"Value": strings.Join(ctxkey.Notifications.Get(c).Supported(), ",")})

			response.InvalidParameterValue(c, "notifications", message, reason, recovery)
			return
		}
	}

	token, err := StorageOf(c).Tokens.GetTokenDetailsByID(c, tokenID)
	if err != nil {
		response.SendError(c, err)
		return
	}

	userID := ctxUser.GetID(c)
	if token.UserID != userID {
		response.NotFound(c)
		return
	}

	if err := StorageOf(c).Tokens.UpdateTokenByID(c, tokenID, name, jsonBody.EnableNotification, subscribedNotifications); err != nil {
		response.InternalServerError(c)
		return
	}

	response.OkWithStatus(c, t("status.auth_token.updated"))
}

// RefreshAuthToken refreshes a auth token.
func RefreshAuthToken(c *gin.Context) {
	tokenID := ctxAuthToken.GetID(c)

	signedToken, err := ctxAuthToken.HooksOf(c).RefreshAuthToken(c, ctxkey.Config.Get(c).JWT.Secret, ctxkey.Config.Get(c).JWT.SigningMethod, tokenID, time.Duration(ctxkey.Config.Get(c).JWT.TTL))
	if err != nil {
		locale := ctxLocale.Get(c)
		t := ctxkey.Localization.Get(c).Translate(locale)

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

// upgradePassword encrypts the password of the user again if its hash is weaker than the current one. The failure is
// logged: the user can sign in anyway.
func upgradePassword(c *gin.Context, userID, password, encryptedPassword string) {
	if !authentication.NeedsRehash(encryptedPassword) {
		return
	}

	upgraded, err := authentication.EncryptPassword(password)
	if err == nil {
		err = StorageOf(c).Accounts.UpdateEncryptedPasswordByUserID(c, userID, upgraded)
	}
	if err != nil {
		log.Error("unable to upgrade the encrypted password",
			"user", userID,
			"error", err.Error(),
		)
	}
}
