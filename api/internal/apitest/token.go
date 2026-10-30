package apitest

import (
	"testing"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
)

// SignToken returns a token signed with the secret, valid or expired.
func SignToken(t *testing.T, secret, userID, tokenID string, expired bool) string {
	t.Helper()

	expiration := time.Now().Add(time.Hour)
	if expired {
		expiration = time.Now().Add(-time.Hour)
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": userID,
		"id":  tokenID,
		"exp": expiration.Unix(),
	})

	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("unable to sign the token: %v", err)
	}

	return signed
}
