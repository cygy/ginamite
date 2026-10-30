package validator

import (
	"fmt"
	"time"

	"github.com/cygy/ginamite/common/authentication"
	resty "github.com/go-resty/resty/v2"
)

type googleTokenValidator struct {
	client   *resty.Client
	clientID string
}

// SourceName returns the source name of the token
func (v *googleTokenValidator) SourceName() string {
	return authentication.MethodGoogle
}

// VerifyTokenAndExtractInfos verifies a google access token
func (v *googleTokenValidator) VerifyTokenAndExtractInfos(token string) (TokenInfos, bool) {
	endpoint := fmt.Sprintf("tokeninfo?id_token=%s", token)
	r, err := v.client.R().Get(endpoint)

	if !logResponseToken(token, v, r, err) {
		return TokenInfos{}, false
	}

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

	if !extractClaimsFromJWTToken(token, &tokenInfos, v) ||
		len(tokenInfos.Sub) == 0 ||
		len(tokenInfos.Email) == 0 ||
		!tokenInfos.EmailVerified ||
		tokenInfos.Aud != v.clientID {
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

// NewGoogleTokenValidator returns a TokenValidator for Google OAuth tokens.
func NewGoogleTokenValidator(clientID, APIVersion string, timeOut, retryCount int, debugMode bool) TokenValidator {
	host := fmt.Sprintf("https://www.googleapis.com/oauth2/%s/", APIVersion)

	return &googleTokenValidator{
		client:   apiClient(host, timeOut, retryCount, debugMode),
		clientID: clientID,
	}
}
