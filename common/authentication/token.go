package authentication

import (
	"context"
	"errors"
	"time"

	"github.com/cygy/ginamite/common/log"

	jwt "github.com/golang-jwt/jwt/v5"
)

// Token holds the content of an auth token.
type Token struct {
	UserID         string
	Source         string
	PrivateKey     string
	ExpirationDate time.Time
	CreationDate   time.Time
	Other          map[string]any
}

// GetSignedToken returns a signed JWT with the details of the auth token tokenID.
// It needs the BuildTokenFromID and ExtraPropertiesForTokenWithID hooks.
// Unknown signing methods fall back to HS256.
func (h Hooks) GetSignedToken(ctx context.Context, signingMethod, secret, tokenID string) (string, error) {
	if h.BuildTokenFromID == nil {
		return "", errors.New("undefined function 'BuildTokenFromID'")
	}

	if h.ExtraPropertiesForTokenWithID == nil {
		return "", errors.New("undefined function 'ExtraPropertiesForTokenWithID'")
	}

	token, err := h.BuildTokenFromID(ctx, tokenID)
	if err != nil {
		log.Error("unable to build JWT token",
			"token id", tokenID,
			"signature_method", signingMethod,
			"error", err,
		)
		return "", err
	}

	extraProperties, err := h.ExtraPropertiesForTokenWithID(ctx, tokenID)
	if err != nil {
		log.Error("unable to get extra properties for JWT token",
			"token id", tokenID,
			"signature_method", signingMethod,
			"error", err,
		)
		return "", err
	}

	return SignToken(signingMethod, secret, tokenID, token, extraProperties)
}

// SignToken returns a signed JWT with the details of the auth token tokenID and the extra properties. Unknown signing
// methods fall back to HS256.
func SignToken(signingMethod, secret, tokenID string, token *Token, extraProperties map[string]string) (string, error) {
	// Create the JWT claims
	jwtClaims := jwt.MapClaims{
		"iss": token.Source,
		"sub": token.UserID,
		"exp": token.ExpirationDate.Unix(),
		"iat": token.CreationDate.Unix(),
		"id":  tokenID,
		"up":  time.Now().Unix(),
	}

	for key, value := range token.Other {
		jwtClaims[key] = value
	}
	for key, value := range extraProperties {
		jwtClaims[key] = value
	}

	// Create the JWT token.
	var signingMethodToUse *jwt.SigningMethodHMAC
	switch signingMethod {
	case "HS256":
		signingMethodToUse = jwt.SigningMethodHS256
	case "HS384":
		signingMethodToUse = jwt.SigningMethodHS384
	case "HS512":
		signingMethodToUse = jwt.SigningMethodHS512
	default:
		signingMethodToUse = jwt.SigningMethodHS256
	}

	jwtToken := jwt.NewWithClaims(signingMethodToUse, jwtClaims)

	// Sign and get the final encoded token as a string.
	signedToken, err := jwtToken.SignedString([]byte(secret))
	if err != nil {
		log.Error("unable to sign JWT token", "claims", jwtClaims, "signature_method", signingMethod)
		return "", err
	}

	return signedToken, nil
}

// ParseToken parses and verifies a signed JWT and returns it. Only the HMAC methods HS256, HS384 and HS512 are accepted
// and the token must expire. The error wraps the ones of the jwt package: use errors.Is with jwt.ErrTokenExpired.
func ParseToken(signedToken, secret string) (*jwt.Token, error) {
	token, err := jwt.Parse(signedToken, func(*jwt.Token) (any, error) {
		return []byte(secret), nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg(), jwt.SigningMethodHS384.Alg(), jwt.SigningMethodHS512.Alg()}), jwt.WithExpirationRequired())
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	return token, nil
}

// RefreshAuthToken extends the validity of the auth token tokenID by ttl and
// returns it freshly signed.
func (h Hooks) RefreshAuthToken(ctx context.Context, secret, signingMethod, tokenID string, ttl time.Duration) (string, error) {
	if h.ExtendTokenExpirationDateFromID == nil {
		return "", errors.New("undefined function 'ExtendTokenExpirationDateFromID'")
	}

	// Extends the validity of the token.
	if err := h.ExtendTokenExpirationDateFromID(ctx, tokenID, ttl); err != nil {
		return "", err
	}

	// Sign and get the complete encoded token as a string.
	signedToken, err := h.GetSignedToken(ctx, signingMethod, secret, tokenID)
	if err != nil {
		return "", err
	}

	return signedToken, nil
}

// GetUserID returns the user ID of a JWT, or an empty string if it has none.
func GetUserID(token *jwt.Token) string {
	if token == nil || token.Claims == nil {
		return ""
	}

	return stringClaim(token, "sub")
}

// GetTokenID returns the auth token ID of a JWT, or an empty string if it has none.
func GetTokenID(token *jwt.Token) string {
	if token == nil || token.Claims == nil {
		return ""
	}

	return stringClaim(token, "id")
}

// GetLastUpdate returns the date of the last update of a JWT, or the Unix epoch if it has none.
func GetLastUpdate(token *jwt.Token) time.Time {
	if token == nil || token.Claims == nil {
		return time.Unix(0, 0)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return time.Unix(0, 0)
	}

	switch value := claims["up"].(type) {
	case int64:
		return time.Unix(value, 0)
	case float64:
		return time.Unix(int64(value), 0)
	}

	return time.Unix(0, 0)
}

// stringClaim returns the string claim name of a JWT, or an empty string.
func stringClaim(token *jwt.Token, name string) string {
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return ""
	}

	value, _ := claims[name].(string)

	return value
}
