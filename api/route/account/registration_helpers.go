package account

import (
	"errors"
	"strings"

	"github.com/cygy/ginamite/common/ctxqueue"

	ctxLocale "github.com/cygy/ginamite/api/context/locale"
	"github.com/cygy/ginamite/api/request"
	"github.com/cygy/ginamite/api/response"
	"github.com/cygy/ginamite/common/authentication"
	"github.com/cygy/ginamite/common/ctxconfig"
	registrationError "github.com/cygy/ginamite/common/errors/registration"
	"github.com/cygy/ginamite/common/localization"
	"github.com/cygy/ginamite/common/queue"
	"github.com/cygy/ginamite/common/random"
	req "github.com/cygy/ginamite/common/request"
	res "github.com/cygy/ginamite/common/response"

	"github.com/gin-gonic/gin"
)

func register(c *gin.Context, method string) {
	if FunctionsOf(c).GetUserDetailsByIdentifier == nil {
		c.Error(errors.New("undefined function 'GetUserDetailsByIdentifier'"))
		response.InternalServerError(c)
		return
	}

	if FunctionsOf(c).RegisterByEmailAddress == nil {
		c.Error(errors.New("undefined function 'RegisterByEmailAddress'"))
		response.InternalServerError(c)
		return
	}

	if FunctionsOf(c).RegisterByThirdPartyToken == nil {
		c.Error(errors.New("undefined function 'RegisterByThirdPartyToken'"))
		response.InternalServerError(c)
		return
	}

	if FunctionsOf(c).DoesAccountWithUsernameExist == nil {
		c.Error(errors.New("undefined function 'DoesAccountWithUsernameExist'"))
		response.InternalServerError(c)
		return
	}

	if FunctionsOf(c).DoesAccountWithEmailAddressExist == nil {
		c.Error(errors.New("undefined function 'DoesAccountWithEmailAddressExist'"))
		response.InternalServerError(c)
		return
	}

	var jsonBody struct {
		Username     string                `json:"username"`
		Email        string                `json:"email_address"` // only register with method = 'password'
		Password     string                `json:"password"`      // only register with method = 'password'
		Token        string                `json:"token"`         // only register with method = 'facebook' or method = 'google'
		Source       string                `json:"source"`
		Device       authentication.Device `json:"device"`
		Locale       string                `json:"locale"`
		TermsVersion string                `json:"terms_version"`
	}
	request.DecodeBody(c, &jsonBody)

	isPasswordMethod := (method == authentication.MethodPassword)
	isFacebookMethod := (method == authentication.MethodFacebook)
	isGoogleMethod := (method == authentication.MethodGoogle)
	isThirdPartyTokenMethod := isFacebookMethod || isGoogleMethod

	// List here the available authentication methods.
	if !isPasswordMethod && !isFacebookMethod && !isGoogleMethod {
		c.Error(errors.New("wrong authentication method '" + method + "'"))
		response.InternalServerError(c)
		return
	}

	locale := ctxLocale.Get(c)
	t := localization.Translate(locale)

	// Verify that the username follows some rules.
	username := jsonBody.Username
	if !CheckUsernameParameter(c, username, "username") {
		return
	}

	errorMessageKey := "error.registration.message"

	emailAddress := jsonBody.Email
	password := jsonBody.Password
	var encryptedPassword string

	if isPasswordMethod {
		if !validateEmailAddress(c, emailAddress, errorMessageKey) || !validatePassword(c, password, errorMessageKey) {
			return
		}

		var err error
		encryptedPassword, err = authentication.EncryptPassword(password)
		if err != nil {
			response.InternalServerError(c)
			return
		}
	}

	token := jsonBody.Token

	if isThirdPartyTokenMethod {
		// Token is mandatory.
		if len(token) == 0 {
			response.NotFoundParameterValue(c, "token",
				t(errorMessageKey),
				t("error.registration.third_party_service_token.not_found.reason"),
				t("error.registration.third_party_service_token.not_found.recovery"),
			)
			return
		}
	}

	// Source is mandatory.
	source := jsonBody.Source
	if len(source) == 0 {
		response.NotFoundParameterValue(c, "source",
			t(errorMessageKey),
			t("error.registration.source.not_found.reason"),
			t("error.registration.source.not_found.recovery"),
		)
		return
	}

	// Device type is mandatory.
	device := jsonBody.Device
	if len(device.Type) == 0 || len(device.Name) == 0 {
		response.NotFoundParameterValue(c, "device",
			t(errorMessageKey),
			t("error.registration.device.not_found.reason"),
			t("error.registration.device.not_found.recovery"),
		)
		return
	}

	// Terms version is mandatory.
	termsVersion := jsonBody.TermsVersion
	if len(termsVersion) == 0 {
		response.NotFoundParameterValue(c, "terms_version",
			t(errorMessageKey),
			t("error.registration.terms_version.not_found.reason"),
			t("error.registration.terms_version.not_found.recovery"),
		)
		return
	}

	if termsVersion != ctxconfig.Get(c).Terms.Version {
		response.InvalidParameterValue(c, "terms_version",
			t(errorMessageKey),
			t("error.terms.version.invalid.reason"),
			t("error.terms.version.invalid.recovery"),
		)
		return
	}

	// Locale is mandatory.
	accountLocale := strings.ToLower(jsonBody.Locale)
	if len(accountLocale) == 0 {
		response.NotFoundParameterValue(c, "locale",
			t(errorMessageKey),
			t("error.registration.locale.not_found.reason"),
			t("error.registration.locale.not_found.recovery"),
		)
		return
	}

	// Locale not supported.
	if !ctxconfig.Get(c).SupportsLocale(accountLocale) {
		response.UnsupportedParameterLocale(c, locale, accountLocale, "locale", ctxconfig.Get(c).SupportedLocales, errorMessageKey)
		return
	}

	sendAccountAlreadyExistsError := func() {
		response.PreconditionFailedWithError(c, res.Error{
			Domain:   registrationError.Domain,
			Code:     registrationError.AccountAlreadyExists,
			Message:  t(errorMessageKey),
			Reason:   t("error.registration.account.already_exists.reason"),
			Recovery: t("error.registration.account.already_exists.recovery"),
		})
	}

	exists, err := FunctionsOf(c).DoesAccountWithUsernameExist(c, username)
	if exists || err != nil {
		sendAccountAlreadyExistsError()
		return
	}

	isEmailAddressVerified := false
	registrationCode := random.String(authentication.RegisterCodeLength)
	privateKey := random.String(authentication.PrivateKeyLength)
	mustBeLoggedin := !ctxconfig.Get(c).Account.EmailAddressMustBeConfirmed
	IPAddress := req.GetRealIPAddress(c)

	// Register with a third party service
	if isThirdPartyTokenMethod {
		// Create a validator to verify the token and extract informations about the user.
		tokenValidator := newTokenValidator(c, method)

		// First verify if the token is valid, invalid tokens are not accepted of course.
		tokenInfos, ok := tokenValidator.VerifyTokenAndExtractInfos(token)
		if !ok {
			response.PreconditionFailedWithError(c, res.Error{
				Domain:   registrationError.Domain,
				Code:     registrationError.TokenCheckFailed,
				Message:  t(errorMessageKey),
				Reason:   t("error.registration.third_party_service_token.can_not_verify.reason"),
				Recovery: t("error.registration.third_party_service_token.can_not_verify.recovery"),
			})
			return
		}

		// Then verify that an exiting account is not linked to the third-party user.
		var DoesUserWithSocialNetworkUserIDExist DoesUserWithSocialNetworkUserIDExistFunc
		if isFacebookMethod {
			DoesUserWithSocialNetworkUserIDExist = FunctionsOf(c).DoesUserWithFacebookUserIDExist
		} else if isGoogleMethod {
			DoesUserWithSocialNetworkUserIDExist = FunctionsOf(c).DoesUserWithGoogleUserIDExist
		}
		if DoesUserWithSocialNetworkUserIDExist != nil {
			exists, err := DoesUserWithSocialNetworkUserIDExist(c, tokenInfos.UserID)
			if exists || err != nil {
				return
			}
		}

		// Finally verify that an account with this email address does not exist.
		exists, err := FunctionsOf(c).DoesAccountWithEmailAddressExist(c, tokenInfos.Email)
		if exists || err != nil {
			sendAccountAlreadyExistsError()
			return
		}

		// Create an account with this third-party token.
		if err := FunctionsOf(c).RegisterByThirdPartyToken(c, username, tokenInfos, tokenValidator.SourceName(), accountLocale, termsVersion, registrationCode, privateKey, source, IPAddress, device); err != nil {
			response.InternalServerError(c)
			return
		}

		// Here, the user ID and the email address are available to be used on a new account.
		isEmailAddressVerified = true
		emailAddress = tokenInfos.Email
	}

	// Register with a email/password
	if isPasswordMethod {
		// Verify that an account with this email address does not exist.
		exists, err := FunctionsOf(c).DoesAccountWithEmailAddressExist(c, emailAddress)
		if exists || err != nil {
			sendAccountAlreadyExistsError()
			return
		}

		// Create an account with this email address.
		if err := FunctionsOf(c).RegisterByEmailAddress(c, username, encryptedPassword, emailAddress, accountLocale, termsVersion, registrationCode, privateKey, source, IPAddress, device); err != nil {
			response.InternalServerError(c)
			return
		}
	}

	// Get the newly created account.
	user, err := FunctionsOf(c).GetUserDetailsByIdentifier(c, emailAddress)
	if err != nil {
		response.InternalServerError(c)
		return
	}

	r := response.H{}

	// Create an authentication token if the email address must not be verified.
	if mustBeLoggedin {
		authToken, _ := createAuthenticationToken(c, user, IPAddress, source, device, method, false, true)
		r["auth_token"] = authToken
	}

	if isEmailAddressVerified {
		// If the email address is valid, a welcome email is sent.
		ctxqueue.Get(c).SendMail(queue.MessageMailRegistrationWelcome, queue.MailRegistrationWelcome{
			EmailAddress: user.EmailAddress,
			Username:     user.Username,
			Locale:       user.Locale,
		})

		r["status"] = t("status.registration.created")
	} else {
		// The email address must be validated by the user.
		// In Env test, we return the code and the user ID.
		if ctxconfig.Get(c).IsTestEnvironment() {
			r["key"] = registrationCode
			r["id"] = user.ID
		} else {
			ctxqueue.Get(c).SendMail(queue.MessageMailRegistrationConfirmation, queue.MailRegistrationConfirmation{
				EmailAddress:     user.EmailAddress,
				UserID:           user.ID,
				Username:         user.Username,
				Locale:           user.Locale,
				RegistrationCode: registrationCode,
			})
		}

		r["status"] = t("status.registration.email_sent")
	}

	ctxqueue.Get(c).RegistrationDone(queue.TaskRegistrationDone{
		UserID: user.ID,
	})

	response.Created(c, r)
}

func validateOrCancelRegistration(c *gin.Context, validate bool) {
	locale := ctxLocale.Get(c)
	t := localization.Translate(locale)

	// Here is a local function that returns an error. This route does only return a single error.
	sendError := func() {
		var message string
		if validate {
			message = t("error.registration.validate.message")
		} else {
			message = t("error.registration.cancel.message")
		}

		response.InvalidRequestParameterWithDetails(c,
			message,
			t("error.registration.validate_or_cancel.reason"),
			t("error.registration.validate_or_cancel.recovery"),
		)
	}

	if FunctionsOf(c).GetUserDetailsByID == nil {
		c.Error(errors.New("undefined function 'GetUserDetailsByID'"))
		sendError()
		return
	}

	if FunctionsOf(c).GetRegistrationDetails == nil {
		c.Error(errors.New("undefined function 'GetRegistrationDetails'"))
		sendError()
		return
	}

	if FunctionsOf(c).ValidateUserRegistration == nil {
		c.Error(errors.New("undefined function 'ValidateUserRegistration'"))
		sendError()
		return
	}

	if FunctionsOf(c).CancelUserRegistration == nil {
		c.Error(errors.New("undefined function 'CancelUserRegistration'"))
		sendError()
		return
	}

	// ID and key are mandatory.
	userID := c.Param("id")
	key := c.Param("key")
	if len(userID) == 0 || len(key) == 0 {
		sendError()
		return
	}

	user, err := FunctionsOf(c).GetUserDetailsByID(c, userID)
	if err != nil {
		sendError()
		return
	}

	var registrationDetails *authentication.Details
	if validate {
		registrationDetails, err = FunctionsOf(c).GetRegistrationDetails(c, userID)
		if err != nil {
			sendError()
			return
		}

		err = FunctionsOf(c).ValidateUserRegistration(c, userID, key)
	} else {
		err = FunctionsOf(c).CancelUserRegistration(c, userID, key)
	}
	if err != nil {
		sendError()
		return
	}

	// Send mail.
	if validate {
		ctxqueue.Get(c).SendMail(queue.MessageMailRegistrationConfirmed, queue.MailRegistrationConfirmed{
			EmailAddress: user.EmailAddress,
			Username:     user.Username,
			Locale:       user.Locale,
		})

		ctxqueue.Get(c).RegistrationValidated(queue.TaskRegistrationValidated{
			UserID: user.ID,
		})

		IPAddress := req.GetRealIPAddress(c)
		// The errors are already reported to the client.
		if _, err := createAuthenticationToken(c, user, IPAddress, registrationDetails.Source, registrationDetails.Device, registrationDetails.Method, true, true); err != nil {
			return
		}
	} else {
		ctxqueue.Get(c).SendMail(queue.MessageMailRegistrationCancelled, queue.MailRegistrationCancelled{
			EmailAddress: user.EmailAddress,
			Username:     user.Username,
			Locale:       user.Locale,
		})
		response.OkWithStatus(c, t("status.registration.cancelled"))
	}
}
