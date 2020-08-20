package account

import (
	"errors"

	"github.com/cygy/ginamite/api/context"
	"github.com/cygy/ginamite/api/request"
	"github.com/cygy/ginamite/api/response"
	"github.com/cygy/ginamite/api/validator"
	"github.com/cygy/ginamite/common/authentication"
	commonErrors "github.com/cygy/ginamite/common/errors"
	authError "github.com/cygy/ginamite/common/errors/auth"
	"github.com/cygy/ginamite/common/localization"
	req "github.com/cygy/ginamite/common/request"
	res "github.com/cygy/ginamite/common/response"

	"github.com/gin-gonic/gin"
)

func loginWithThirdParty(c *gin.Context, v validator.TokenValidator) {
	var jsonBody struct {
		Token  string                `json:"token"`
		Source string                `json:"source"`
		Device authentication.Device `json:"device"`
	}
	request.DecodeBody(c, &jsonBody)

	authenticationMethod := v.SourceName()
	isFacebookMethod := (authenticationMethod == authentication.MethodFacebook)
	isGoogleMethod := (authenticationMethod == authentication.MethodGoogle)

	locale := context.GetLocale(c)
	t := localization.Translate(locale)

	errorMessageKey := "error.login.unable.message"

	// Token is mandatory.
	token := jsonBody.Token
	if len(token) == 0 {
		response.NotFoundParameterValue(c, "token",
			t(errorMessageKey),
			t("error.login.third_party_service_token.not_found.reason"),
			t("error.login.third_party_service_token.not_found.recovery"),
		)
		return
	}

	// Source is mandatory.
	source := jsonBody.Source
	if len(source) == 0 {
		response.NotFoundParameterValue(c, "source",
			t(errorMessageKey),
			t("error.login.source.not_found.reason"),
			t("error.login.source.not_found.recovery"),
		)
		return
	}

	// Device is mandatory.
	device := jsonBody.Device
	if len(device.Type) == 0 || len(device.Name) == 0 {
		response.NotFoundParameterValue(c, "device",
			t(errorMessageKey),
			t("error.login.device.not_found.reason"),
			t("error.login.device.not_found.recovery"),
		)
		return
	}

	// Verify the token received
	tokenInfos, ok := v.VerifyTokenAndExtractInfos(token)
	if !ok {
		canNotLoginError(c, nil)
		return
	}

	// Get the user document, first by the userID, then by the email address.
	var user authentication.User
	var err error
	if isFacebookMethod {
		if GetUserDetailsByFacebookUserID == nil {
			err := errors.New("undefined function 'GetUserDetailsByFacebookUserID'")
			canNotLoginError(c, err)
			return
		}
		user, err = GetUserDetailsByFacebookUserID(c, tokenInfos.UserID)
	} else if isGoogleMethod {
		if GetUserDetailsByGoogleUserID == nil {
			err := errors.New("undefined function 'GetUserDetailsByGoogleUserID'")
			canNotLoginError(c, err)
			return
		}
		user, err = GetUserDetailsByGoogleUserID(c, tokenInfos.UserID)
	}
	if err != nil && !commonErrors.IsNotFound(err) {
		canNotLoginError(c, err)
		return
	}

	if !user.IsSaved() {
		if GetUserDetailsByIdentifier == nil {
			err := errors.New("undefined function 'GetUserDetailsByIdentifier'")
			canNotLoginError(c, err)
			return
		}

		if user, err = GetUserDetailsByIdentifier(c, tokenInfos.Email); err != nil && !commonErrors.IsNotFound(err) {
			canNotLoginError(c, err)
			return
		}
	}

	if !user.IsSaved() {
		// Tell the user that no account was found, the user must create one.
		response.NotFound(c)
		return
	}

	// Save the token.
	if isFacebookMethod {
		if SaveFacebookUserTokenDetailsToUser == nil {
			err := errors.New("undefined function 'SaveFacebookUserTokenDetailsToUser'")
			canNotLoginError(c, err)
			return
		}
		err = SaveFacebookUserTokenDetailsToUser(c, user.ID, tokenInfos)
	} else if isGoogleMethod {
		if SaveGoogleUserTokenDetailsToUser == nil {
			err := errors.New("undefined function 'SaveGoogleUserTokenDetailsToUser'")
			canNotLoginError(c, err)
			return
		}
		err = SaveGoogleUserTokenDetailsToUser(c, user.ID, tokenInfos)
	}
	if err != nil {
		canNotLoginError(c, err)
		return
	}

	IPAddress := req.GetRealIPAddress(c)
	createAuthenticationToken(c, user, IPAddress, source, device, authenticationMethod, true, false)
}

func canNotLoginError(c *gin.Context, err error) {
	if err != nil {
		c.Error(err)
	}

	locale := context.GetLocale(c)
	t := localization.Translate(locale)

	response.BadRequestWithError(c, res.Error{
		Domain:   authError.Domain,
		Code:     authError.CanNotLogin,
		Message:  t("error.login.unable.message"),
		Reason:   t("error.login.unable.reason"),
		Recovery: t("error.login.unable.recovery"),
	})
}
