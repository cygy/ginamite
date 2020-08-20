package authentication

import (
	"errors"
	"fmt"
	"time"

	"github.com/cygy/ginamite/common/log"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/sirupsen/logrus"
)

// GetSignedToken : returns a JWT token with some details of an user and an auth token.
func GetSignedToken(signingMethod, secret, tokenID string) (string, error) {
	if BuildTokenFromID == nil {
		return "", errors.New("undefined function 'BuildTokenFromID'")
	}

	if ExtraPropertiesForTokenWithID == nil {
		return "", errors.New("undefined function 'ExtraPropertiesForTokenWithID'")
	}

	// Get the token details.
	var token *Token
	var extraProperties map[string]string
	var err error

	if tokenID == "fakeexpiredtoken" {
		token = &Token{
			Source:         "debug/test",
			UserID:         "ac006c77fd78cff13e0fa9d1",
			ExpirationDate: time.Now().Add(time.Second * -1000),
			CreationDate:   time.Now().Add(time.Second * -10000),
		}
		extraProperties = map[string]string{}
	} else {
		token, err = BuildTokenFromID(tokenID)
		if err != nil {
			log.WithFields(logrus.Fields{
				"token id":         tokenID,
				"signature_method": signingMethod,
				"error":            err,
			}).Error("unable to build JWT token")
			return "", err
		}

		extraProperties, err = ExtraPropertiesForTokenWithID(tokenID)
		if err != nil {
			log.WithFields(logrus.Fields{
				"token id":         tokenID,
				"signature_method": signingMethod,
				"error":            err,
			}).Error("unable to get extra properties for JWT token")
			return "", err
		}
	}

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
		log.WithFields(logrus.Fields(jwtClaims)).WithField("signature_method", signingMethod).Error("unable to sign JWT token")
		return "", err
	}

	return signedToken, nil
}

// ParseToken : parses and verifies the authentication token and returns the claims.
func ParseToken(signedToken, secret string) (*jwt.Token, error) {
	token, err := jwt.Parse(signedToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil || !token.Valid {
		return nil, err
	}

	return token, nil
}

// RefreshAuthToken : refreshes and returns an authentication token.
func RefreshAuthToken(secret, signingMethod, tokenID string, ttl time.Duration) (string, error) {
	if ExtendTokenExpirationDateFromID == nil {
		return "", errors.New("undefined function 'ExtendTokenExpirationDateFromID'")
	}

	// Extends the validity of the token.
	if err := ExtendTokenExpirationDateFromID(tokenID, ttl); err != nil {
		return "", err
	}

	// Sign and get the complete encoded token as a string.
	signedToken, err := GetSignedToken(signingMethod, secret, tokenID)
	if err != nil {
		return "", err
	}

	return signedToken, nil
}

// GetUserID : returns the user ID from a JWT token.
func GetUserID(token *jwt.Token) string {
	if token == nil || token.Claims == nil {
		return ""
	}

	if value, ok := token.Claims.(jwt.MapClaims)["sub"]; ok {
		return value.(string)
	}

	return ""
}

// GetTokenID : returns the token ID from a JWT token.
func GetTokenID(token *jwt.Token) string {
	if token == nil || token.Claims == nil {
		return ""
	}

	if value, ok := token.Claims.(jwt.MapClaims)["id"]; ok {
		return value.(string)
	}

	return ""
}

// GetLastUpdate : returns the token ID from a JWT token.
func GetLastUpdate(token *jwt.Token) time.Time {
	if token == nil || token.Claims == nil {
		return time.Unix(0, 0)
	}

	if value, ok := token.Claims.(jwt.MapClaims)["up"]; ok {
		return time.Unix(value.(int64), 0)
	}

	return time.Unix(0, 0)
}
