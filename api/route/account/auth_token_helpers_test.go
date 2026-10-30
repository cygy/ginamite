package account

import (
	"net/http"
	"testing"
	"time"

	"github.com/cygy/ginamite/common/authentication"
	commonErrors "github.com/cygy/ginamite/common/errors"

	"github.com/gin-gonic/gin"
)

// deleteStubs defines the functions used by deleteAuthenticationToken.
func deleteStubs(token authentication.Token, getErr, deleteErr error) func() {
	return func() {
		GetTokenDetailsByID = func(*gin.Context, string) (authentication.Token, error) { return token, getErr }
		DeleteTokenByID = func(*gin.Context, string) error { return deleteErr }
	}
}

func TestDeleteAuthenticationToken(t *testing.T) {
	token := authentication.Token{UserID: testUserID, PrivateKey: "token-key", ExpirationDate: time.Now().Add(time.Hour)}

	t.Run("without key", func(t *testing.T) {
		runCases(t, DeleteAuthToken, []testCase{
			{name: "GetTokenDetailsByID not defined", userID: testUserID, id: "t1", setup: func() {
				DeleteTokenByID = func(*gin.Context, string) error { return nil }
			}, want: http.StatusInternalServerError},
			{name: "DeleteTokenByID not defined", userID: testUserID, id: "t1", setup: func() {
				GetTokenDetailsByID = func(*gin.Context, string) (authentication.Token, error) { return token, nil }
			}, want: http.StatusNotFound},
			{name: "token not found", userID: testUserID, id: "t1", setup: deleteStubs(token, commonErrors.NotFound(), nil), want: http.StatusNotFound},
			{name: "lookup error", userID: testUserID, id: "t1", setup: deleteStubs(token, errBoom, nil), want: http.StatusInternalServerError},
			{name: "deletion error", userID: testUserID, id: "t1", setup: deleteStubs(token, nil, errBoom), want: http.StatusInternalServerError},
		})
	})

	t.Run("with key", func(t *testing.T) {
		runCases(t, DeleteAuthTokenWithKey, []testCase{
			{name: "token not found", id: "t1", key: "token-key", setup: deleteStubs(token, commonErrors.NotFound(), nil), want: http.StatusBadRequest},
			{name: "DeleteTokenByID not defined", id: "t1", key: "token-key", setup: func() {
				GetTokenDetailsByID = func(*gin.Context, string) (authentication.Token, error) { return token, nil }
			}, want: http.StatusBadRequest},
		})
	})
}

func TestCreateAuthenticationToken(t *testing.T) {
	device := authentication.Device{Type: "mobile", Name: "iPhone"}
	tests := []struct {
		name         string
		setup        func()
		afterAccount bool
		sendResponse bool
		wantErr      bool
		wantStatus   int
	}{
		{"SaveAuthenticationToken not defined", func() { DeleteTokenByID = func(*gin.Context, string) error { return nil } }, false, true, true, http.StatusBadRequest},
		{"DeleteTokenByID not defined", func() {
			SaveAuthenticationToken = func(*gin.Context, authentication.Details) (string, bool, error) { return testTokenID, true, nil }
		}, false, true, true, http.StatusBadRequest},
		{"save error", all(loginStubs, func() {
			SaveAuthenticationToken = func(*gin.Context, authentication.Details) (string, bool, error) { return "", false, errBoom }
		}), false, true, true, http.StatusBadRequest},
		{"sign error", all(loginStubs, func() {
			authentication.BuildTokenFromID = nil
		}), false, true, true, http.StatusBadRequest},
		{"sign error and deletion error", all(loginStubs, func() {
			authentication.BuildTokenFromID = nil
			DeleteTokenByID = func(*gin.Context, string) error { return errBoom }
		}), false, true, true, http.StatusBadRequest},
		{"success", loginStubs, false, true, false, http.StatusCreated},
		{"success with known IP address", all(loginStubs, func() {
			SaveAuthenticationToken = func(*gin.Context, authentication.Details) (string, bool, error) { return testTokenID, true, nil }
		}), false, true, false, http.StatusCreated},
		{"success after validating the account", loginStubs, true, true, false, http.StatusOK},
		{"success in test environment", all(loginStubs, inTestEnvironment), false, true, false, http.StatusCreated},
		{"success without response", loginStubs, false, false, false, http.StatusOK},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setup(t)
			tt.setup()

			c, recorder := newContext(nil)
			token, err := createAuthenticationToken(c, testUser, "127.0.0.1", "web", device, authentication.MethodPassword, tt.sendResponse, tt.afterAccount)
			c.Writer.WriteHeaderNow()

			if (err != nil) != tt.wantErr {
				t.Fatalf("error = %v, want error: %v", err, tt.wantErr)
			}
			if !tt.wantErr && token == "" {
				t.Error("token is empty, want a signed token")
			}
			if recorder.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", recorder.Code, tt.wantStatus)
			}
		})
	}
}
