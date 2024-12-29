package validator

import (
	"fmt"
	"time"

	"github.com/cygy/ginamite/common/authentication"

	resty "github.com/go-resty/resty/v2"
)

// GoogleTokenValidator to validate the google OAuth tokens
type GoogleTokenValidator struct {
	APIClient *resty.Client
	ClientID  string
}

// NewGoogleTokenValidator returns a new GoogleTokenValidator struct
func NewGoogleTokenValidator(clientID, APIVersion string, timeOut, retryCount int, debugMode bool) *GoogleTokenValidator {
	host := fmt.Sprintf("https://www.googleapis.com/oauth2/%s/", APIVersion)

	return &GoogleTokenValidator{
		APIClient: apiClient(host, timeOut, retryCount, debugMode),
		ClientID:  clientID,
	}
}

// SourceName returns the source name of the token
func (validator *GoogleTokenValidator) SourceName() string {
	return authentication.MethodGoogle
}

// VerifyTokenAndExtractInfos verifies a facebook access token
func (validator *GoogleTokenValidator) VerifyTokenAndExtractInfos(token string) (TokenInfos, bool) {
	// Verify the token is valid
	endpoint := fmt.Sprintf("tokeninfo?id_token=%s", token)
	r, err := validator.APIClient.R().Get(endpoint)

	if !logResponseToken(token, validator, r, err) {
		return TokenInfos{}, false
	}

	// Parse the token to get the informations
	var tokenInfos struct {
		Aud           string
		Sub           string
		GivenName     string `json:"given_name"`
		FamilyName    string `json:"family_name"`
		Email         string
		EmailVerified bool `json:"email_verified"`
		Iat           int64
		Exp           int64
	}

	if !extractClaimsFromJWTToken(token, &tokenInfos, validator) ||
		len(tokenInfos.Sub) == 0 ||
		len(tokenInfos.Email) == 0 ||
		!tokenInfos.EmailVerified ||
		tokenInfos.Aud != validator.ClientID {
		return TokenInfos{}, false
	}

	t := TokenInfos{
		Token:     token,
		UserID:    tokenInfos.Sub,
		Email:     tokenInfos.Email,
		FirstName: tokenInfos.GivenName,
		LastName:  tokenInfos.FamilyName,
		CreatedAt: time.Unix(tokenInfos.Iat, 0),
		ExpiresAt: time.Unix(tokenInfos.Exp, 0),
	}

	return t, true
}
