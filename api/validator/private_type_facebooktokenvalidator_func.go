package validator

import (
	"fmt"
	"time"

	"github.com/cygy/ginamite/common/authentication"
)

// SourceName returns the source name of the token
func (v *facebookTokenValidator) SourceName() string {
	return authentication.MethodFacebook
}

// VerifyTokenAndExtractInfos verifies a facebook access token and extracts the user informations from the token
func (v *facebookTokenValidator) VerifyTokenAndExtractInfos(token string) (TokenInfos, bool) {
	var tokenInfos struct {
		Data struct {
			AppID     string `json:"app_id"`
			ExpiresAt int64  `json:"expires_at"`
			IsValid   bool   `json:"is_valid"`
		}
	}

	endpoint := fmt.Sprintf("debug_token?input_token=%s&access_token=%s", token, token)
	r, err := v.client.R().SetAuthToken(token).Get(endpoint)

	if !logResponseToken(token, v, r, err) ||
		!logParseToken(token, v, r.Body(), &tokenInfos) ||
		!tokenInfos.Data.IsValid ||
		tokenInfos.Data.AppID != v.applicationID {
		return TokenInfos{}, false
	}

	var userInfos struct {
		ID        string `json:"id"`
		Email     string `json:"email"`
		LastName  string `json:"last_name"`
		FirstName string `json:"first_name"`
	}

	endpoint = "me?fields=id,email,first_name,last_name"
	r, err = v.client.R().SetAuthToken(token).Get(endpoint)

	if !logResponseToken(token, v, r, err) ||
		!logParseToken(token, v, r.Body(), &userInfos) {
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
