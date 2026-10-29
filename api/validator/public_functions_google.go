package validator

import "fmt"

// NewGoogleTokenValidator : returns a TokenValidator for Google OAuth tokens.
func NewGoogleTokenValidator(clientID, APIVersion string, timeOut, retryCount int, debugMode bool) TokenValidator {
	host := fmt.Sprintf("https://www.googleapis.com/oauth2/%s/", APIVersion)

	return &googleTokenValidator{
		client:   apiClient(host, timeOut, retryCount, debugMode),
		clientID: clientID,
	}
}
