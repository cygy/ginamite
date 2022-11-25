package context

import (
	"github.com/cygy/ginamite/api/cache"
	"github.com/cygy/ginamite/api/response"
	"github.com/cygy/ginamite/common/authentication"
	"github.com/cygy/ginamite/common/log"

	"github.com/gin-gonic/gin"
	jwt "github.com/golang-jwt/jwt"
	"github.com/golang-jwt/jwt/request"
	"github.com/sirupsen/logrus"
)

// GetUserAndLocaleFromAuthTokenFunc : returns a struct represnting the user and its loale from an auth token.
type GetUserAndLocaleFromAuthTokenFunc func(token *jwt.Token) (interface{}, string)

// IsAuthTokenValidFunc : returns trus if the auth token is still valid and not expired or revoked.
type IsAuthTokenValidFunc func(c *gin.Context, tokenID string) bool

// ParseAuthToken : parses and saves the authenticated user to the context.
func ParseAuthToken(secret string, buildUser GetUserAndLocaleFromAuthTokenFunc, authTokenMustBePresent bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		requestLocale := GetRequestLocale(c)

		// Extract the JWT token.
		authTokenIsPresent := true
		signedToken, err := request.OAuth2Extractor.ExtractToken(c.Request)
		if err != nil {
			if err.Error() == "no token present in request" {
				authTokenIsPresent = false
			}

			if authTokenMustBePresent || authTokenIsPresent {
				log.WithFields(logrus.Fields{
					"error": err.Error(),
				}).Error("unable to extract the authentication token from the headers")
				response.UnauthorizedWithInvalidAuthorizationToken(c, requestLocale)
				return
			}
		}

		var definedLocale string

		if authTokenIsPresent {
			// Verify the JWT token.
			token, err := authentication.ParseToken(signedToken, secret)
			if token == nil {
				log.WithFields(logrus.Fields{
					"token": signedToken,
					"error": err.Error(),
				}).Error("unable to validate the authentication token from the headers")

				if err.Error() == "Token is expired" {
					response.UnauthorizedWithExpiredAuthorizationToken(c, requestLocale)
				} else {
					response.UnauthorizedWithInvalidAuthorizationToken(c, requestLocale)
				}
				return
			}

			// Saves the auth token to the context.
			c.Set(authTokenContextKey, signedToken)
			c.Set(authTokenParsedContextKey, token)

			// Verify if the token is revoked.
			tokenID := GetAuthTokenID(c)
			if cache.IsRevokedAuthToken(tokenID) {
				response.UnauthorizedWithRevokedAuthorizationToken(c, requestLocale)
				return
			}

			// Build the user from the token.
			if buildUser != nil {
				var user interface{}
				user, definedLocale = buildUser(token)
				c.Set(userContextKey, user)
			}
		}

		// Define the user locale.
		userLocale := ""
		if isLocaleSupported(definedLocale) {
			userLocale = definedLocale
		}

		// Save the locales to the context.
		setLocales(c, userLocale, requestLocale)
	}
}

// VerifyAuthToken : checks the auth token is valid and not expired/revoked.
func VerifyAuthToken(verifyToken IsAuthTokenValidFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenID := GetAuthTokenID(c)
		if verifyToken != nil && !verifyToken(c, tokenID) {
			requestLocale := GetRequestLocale(c)
			response.UnauthorizedWithRevokedAuthorizationToken(c, requestLocale)
		}
	}
}
