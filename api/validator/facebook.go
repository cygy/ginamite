package validator

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/cygy/ginamite/common/authentication"
	resty "github.com/go-resty/resty/v2"
)

type facebookTokenValidator struct {
	client            *resty.Client
	applicationID     string
	applicationSecret string
}

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

	// The token is inspected with the access token of the application, which only the API knows.
	r, err := v.client.R().SetQueryParams(map[string]string{
		"input_token":  token,
		"access_token": v.applicationID + "|" + v.applicationSecret,
	}).Get("debug_token")

	if !logResponseToken(v, r, err) ||
		!logParseToken(v, r.Body(), &tokenInfos) ||
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

	// The proof shows that the call comes from the application, not from someone who stole the token.
	r, err = v.client.R().SetAuthToken(token).SetQueryParams(map[string]string{
		"fields":          "id,email,first_name,last_name",
		"appsecret_proof": appSecretProof(token, v.applicationSecret),
	}).Get("me")

	if !logResponseToken(v, r, err) ||
		!logParseToken(v, r.Body(), &userInfos) {
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

// appSecretProof returns the proof of the calls to the Graph API made with the token: the HMAC-SHA256 of the token keyed
// by the secret of the application.
func appSecretProof(token, applicationSecret string) string {
	mac := hmac.New(sha256.New, []byte(applicationSecret))
	mac.Write([]byte(token))

	return hex.EncodeToString(mac.Sum(nil))
}

// NewFacebookTokenValidator returns a TokenValidator for Facebook OAuth tokens. The ID and the secret of the
// application verify that the tokens were issued for it.
func NewFacebookTokenValidator(applicationID, applicationSecret, APIVersion string, timeOut, retryCount int, debugMode bool) TokenValidator {
	host := fmt.Sprintf("https://graph.facebook.com/%s/", APIVersion)

	return &facebookTokenValidator{
		client:            apiClient(host, timeOut, retryCount, debugMode),
		applicationID:     applicationID,
		applicationSecret: applicationSecret,
	}
}
