package html

import (
	"github.com/cygy/ginamite/common/authentication"

	"github.com/gin-gonic/gin"
	jwt "github.com/golang-jwt/jwt/v5"
)

// InitializeData resets the content of the page: it defines the title, the variables of the page and the user found
// in the JWT cookie of the request, whose timezone overrides the one of the page.
// An invalid token, or a token with missing claims, is ignored.
func (page *Page) InitializeData(c *gin.Context) {
	page.Content = PageContent{}
	page.Content.Title = page.Title

	for key, value := range page.Vars {
		page.SetVar(key, value)
	}

	tokenString, err := c.Cookie(jwtCookieName)
	if err != nil {
		return
	}

	token, err := authentication.ParseToken(tokenString, jwtSecret)
	if err != nil || token == nil {
		return
	}

	claims, _ := token.Claims.(jwt.MapClaims)

	page.Content.CurrentUser.ID = stringClaim(claims, "sub")
	page.Content.CurrentUser.Name = stringClaim(claims, "nickname")
	page.Content.CurrentUser.Image1x = stringClaim(claims, "img1")
	page.Content.CurrentUser.Image2x = stringClaim(claims, "img2")
	if unread, ok := claims["unread"].(float64); ok && unread > 0 {
		page.Content.CurrentUser.Unread = uint(unread)
	}

	if timezone := stringClaim(claims, "timezone"); len(timezone) > 0 {
		page.Timezone = timezone
	}
}

// stringClaim returns the string claim name, or an empty string if it is missing or not a string.
func stringClaim(claims jwt.MapClaims, name string) string {
	value, _ := claims[name].(string)

	return value
}
