package account

import (
	"testing"

	"github.com/gin-gonic/gin"
)

func TestStartProcessValidateNewEmailAddress(t *testing.T) {
	setup(t)
	c, _ := newContext(nil)

	if _, _, err := startProcessValidateNewEmailAddress(c, testUserID, "john@example.com", "john", "en"); err == nil {
		t.Error("error = nil, want an error when CreateNewVerifyEmailAddressProcess is not defined")
	}

	CreateNewVerifyEmailAddressProcess = func(*gin.Context, string) (string, string, error) { return "", "", errBoom }
	if _, _, err := startProcessValidateNewEmailAddress(c, testUserID, "john@example.com", "john", "en"); err == nil {
		t.Error("error = nil, want the error of the process creation")
	}

	CreateNewVerifyEmailAddressProcess = func(*gin.Context, string) (string, string, error) { return "id", "key", nil }
	id, key, err := startProcessValidateNewEmailAddress(c, testUserID, "john@example.com", "john", "en")
	if err != nil || id != "id" || key != "key" {
		t.Errorf("startProcessValidateNewEmailAddress() = %q, %q, %v, want id, key, nil", id, key, err)
	}
}
