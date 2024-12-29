package validator

import (
	"fmt"
	"time"

	"github.com/cygy/ginamite/common/authentication"

	resty "github.com/go-resty/resty/v2"
)

// FacebookTokenValidator to validate the facebook OAuth tokens
type FacebookTokenValidator struct {
	APIClient     *resty.Client
	ApplicationID string
}

// NewFacebookTokenValidator returns a new FacebookUserInfos struct
func NewFacebookTokenValidator(applicationID, APIVersion string, timeOut, retryCount int, debugMode bool) *FacebookTokenValidator {
	host := fmt.Sprintf("https://graph.facebook.com/%s/", APIVersion)

	return &FacebookTokenValidator{
		APIClient:     apiClient(host, timeOut, retryCount, debugMode),
		ApplicationID: applicationID,
	}
}

// SourceName returns the source name of the token
func (validator *FacebookTokenValidator) SourceName() string {
	return authentication.MethodFacebook
}

// VerifyTokenAndExtractInfos verifies a facebook access token and extracts the user informations from the token
func (validator *FacebookTokenValidator) VerifyTokenAndExtractInfos(token string) (TokenInfos, bool) {
	// Verify the token is valid
	var tokenInfos struct {
		Data struct {
			AppID     string `json:"app_id"`
			ExpiresAt int64  `json:"expires_at"`
			IsValid   bool   `json:"is_valid"`
		}
	}

	endpoint := fmt.Sprintf("debug_token?input_token=%s&access_token=%s", token, token)
	r, err := validator.APIClient.R().SetAuthToken(token).Get(endpoint)

	if !logResponseToken(token, validator, r, err) ||
		!logParseToken(token, validator, r.Body(), &tokenInfos) ||
		!tokenInfos.Data.IsValid ||
		tokenInfos.Data.AppID != validator.ApplicationID {
		return TokenInfos{}, false
	}

	// Get the informations about the user
	var userInfos struct {
		ID        string `json:"id"`
		Email     string `json:"email"`
		LastName  string `json:"last_name"`
		FirstName string `json:"first_name"`
	}

	endpoint = "me?fields=id,email,first_name,last_name"
	r, err = validator.APIClient.R().SetAuthToken(token).Get(endpoint)

	if !logResponseToken(token, validator, r, err) ||
		!logParseToken(token, validator, r.Body(), &userInfos) {
		return TokenInfos{}, false
	}

	t := TokenInfos{
		Token:     token,
		UserID:    userInfos.ID,
		Email:     userInfos.Email,
		FirstName: userInfos.FirstName,
		LastName:  userInfos.LastName,
		CreatedAt: time.Now(),
		ExpiresAt: time.Unix(tokenInfos.Data.ExpiresAt, 0),
	}

	return t, true
}
