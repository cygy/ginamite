package functions

import (
	"github.com/cygy/ginamite/plugins/mongodb/model"

	jwt "github.com/golang-jwt/jwt/v5"
)

// GetUserAndLocaleFromAuthToken returns the model.User built from the claims of the auth token, and the locale of
// the user. The user is stored by value in the context, see context.GetUser.
func GetUserAndLocaleFromAuthToken(token *jwt.Token) (any, string, error) {
	user, err := model.UserFromAuthToken(token)
	if err != nil {
		return nil, "", err
	}

	return *user, user.Settings.Locale, nil
}
