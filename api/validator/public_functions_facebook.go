package validator

import "fmt"

// NewFacebookTokenValidator : returns a TokenValidator for Facebook OAuth tokens.
func NewFacebookTokenValidator(applicationID, APIVersion string, timeOut, retryCount int, debugMode bool) TokenValidator {
	host := fmt.Sprintf("https://graph.facebook.com/%s/", APIVersion)

	return &facebookTokenValidator{
		client:        apiClient(host, timeOut, retryCount, debugMode),
		applicationID: applicationID,
	}
}
