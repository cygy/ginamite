package authentication

import (
	"context"
	"errors"
	"testing"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
)

const testSecret = "a-very-secret-key"

var errBoom = errors.New("boom")

func okBuild(context.Context, string) (*Token, error) {
	now := time.Now()

	return &Token{
		UserID:         "user-1",
		Source:         "test",
		CreationDate:   now.Add(-time.Minute),
		ExpirationDate: now.Add(time.Hour),
		Other:          map[string]any{"role": "admin"},
	}, nil
}

func okExtra(context.Context, string) (map[string]string, error) {
	return map[string]string{"plan": "pro"}, nil
}

func TestGetSignedToken_Errors(t *testing.T) {
	tests := []struct {
		name  string
		build BuildTokenFunc
		extra ExtraPropertiesForTokenWithIDFunc
		want  error
	}{
		{"undefined build", nil, okExtra, nil},
		{"undefined extra", okBuild, nil, nil},
		{
			"build error",
			func(context.Context, string) (*Token, error) { return nil, errBoom },
			okExtra,
			errBoom,
		},
		{
			"extra error",
			okBuild,
			func(context.Context, string) (map[string]string, error) { return nil, errBoom },
			errBoom,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hooks := Hooks{tt.build, tt.extra, nil}

			signed, err := hooks.GetSignedToken(context.Background(), "HS256", testSecret, "token-1")
			if err == nil || signed != "" {
				t.Fatalf("GetSignedToken() = %q, %v, want an error", signed, err)
			}
			if tt.want != nil && !errors.Is(err, tt.want) {
				t.Errorf("error = %v, want %v", err, tt.want)
			}
		})
	}
}

func TestGetSignedToken_PassesTheContext(t *testing.T) {
	type key struct{}
	ctx := context.WithValue(context.Background(), key{}, "request")

	var got any
	extra := func(ctx context.Context, _ string) (map[string]string, error) {
		got = ctx.Value(key{})
		return nil, nil
	}
	if _, err := (Hooks{okBuild, extra, nil}).GetSignedToken(ctx, "HS256", testSecret, "token-1"); err != nil {
		t.Fatal(err)
	}
	if got != "request" {
		t.Errorf("context value = %v, want the context of the request", got)
	}
}

func TestGetSignedToken(t *testing.T) {
	hooks := Hooks{okBuild, okExtra, nil}

	tests := []struct {
		name   string
		method string
		alg    string
	}{
		{"HS256", "HS256", "HS256"},
		{"HS384", "HS384", "HS384"},
		{"HS512", "HS512", "HS512"},
		{"unknown falls back to HS256", "RS256", "HS256"},
		{"empty falls back to HS256", "", "HS256"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			signed, err := hooks.GetSignedToken(context.Background(), tt.method, testSecret, "token-1")
			if err != nil {
				t.Fatalf("GetSignedToken() error = %v", err)
			}

			token, err := ParseToken(signed, testSecret)
			if err != nil {
				t.Fatalf("ParseToken() error = %v", err)
			}
			if token.Method.Alg() != tt.alg {
				t.Errorf("alg = %s, want %s", token.Method.Alg(), tt.alg)
			}

			claims := token.Claims.(jwt.MapClaims)
			for key, want := range map[string]string{"iss": "test", "sub": "user-1", "id": "token-1", "role": "admin", "plan": "pro"} {
				if claims[key] != want {
					t.Errorf("claim %q = %v, want %q", key, claims[key], want)
				}
			}
			if GetUserID(token) != "user-1" || GetTokenID(token) != "token-1" {
				t.Errorf("GetUserID/GetTokenID = %q/%q", GetUserID(token), GetTokenID(token))
			}
			if time.Since(GetLastUpdate(token)) > time.Minute {
				t.Errorf("GetLastUpdate() = %v, want a recent date", GetLastUpdate(token))
			}
		})
	}
}

func TestSignToken_Expired(t *testing.T) {
	token := &Token{UserID: "u", ExpirationDate: time.Now().Add(-time.Minute), CreationDate: time.Now().Add(-time.Hour)}

	signed, err := SignToken("HS256", testSecret, "t", token, nil)
	if err != nil {
		t.Fatalf("SignToken() error = %v", err)
	}

	if _, err := ParseToken(signed, testSecret); !errors.Is(err, jwt.ErrTokenExpired) {
		t.Errorf("ParseToken() error = %v, want an expired token", err)
	}
}

func TestParseToken(t *testing.T) {
	hooks := Hooks{okBuild, okExtra, nil}

	valid, err := hooks.GetSignedToken(context.Background(), "HS256", testSecret, "token-1")
	if err != nil {
		t.Fatal(err)
	}

	none, err := jwt.NewWithClaims(jwt.SigningMethodNone, jwt.MapClaims{"sub": "x"}).SignedString(jwt.UnsafeAllowNoneSignatureType)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		signed  string
		secret  string
		wantErr bool
	}{
		{"valid", valid, testSecret, false},
		{"wrong secret", valid, "other", true},
		{"garbage", "not.a.jwt", testSecret, true},
		{"empty", "", testSecret, true},
		{"none algorithm", none, testSecret, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := ParseToken(tt.signed, tt.secret)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ParseToken() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr && token != nil {
				t.Error("the token must be nil on error")
			}
		})
	}
}

func TestParseToken_Policy(t *testing.T) {
	sign := func(method jwt.SigningMethod, claims jwt.MapClaims) string {
		t.Helper()

		signed, err := jwt.NewWithClaims(method, claims).SignedString([]byte(testSecret))
		if err != nil {
			t.Fatal(err)
		}

		return signed
	}
	future := time.Now().Add(time.Hour).Unix()

	for _, method := range []jwt.SigningMethod{jwt.SigningMethodHS256, jwt.SigningMethodHS384, jwt.SigningMethodHS512} {
		if _, err := ParseToken(sign(method, jwt.MapClaims{"exp": future}), testSecret); err != nil {
			t.Errorf("%s: %v", method.Alg(), err)
		}
	}

	_, err := ParseToken(sign(jwt.SigningMethodHS256, jwt.MapClaims{"exp": time.Now().Add(-time.Hour).Unix()}), testSecret)
	if !errors.Is(err, jwt.ErrTokenExpired) {
		t.Errorf("expired token: error = %v, want jwt.ErrTokenExpired", err)
	}

	if _, err := ParseToken(sign(jwt.SigningMethodHS256, jwt.MapClaims{"sub": "no-expiration"}), testSecret); err == nil {
		t.Error("a token without expiration must be refused")
	}
}

func TestRefreshAuthToken(t *testing.T) {
	t.Run("undefined extend", func(t *testing.T) {
		hooks := Hooks{okBuild, okExtra, nil}

		if _, err := hooks.RefreshAuthToken(context.Background(), testSecret, "HS256", "token-1", time.Hour); err == nil {
			t.Error("expected an error")
		}
	})

	t.Run("extend error", func(t *testing.T) {
		hooks := Hooks{okBuild, okExtra, func(context.Context, string, time.Duration) error { return errBoom }}

		if _, err := hooks.RefreshAuthToken(context.Background(), testSecret, "HS256", "token-1", time.Hour); !errors.Is(err, errBoom) {
			t.Errorf("error = %v, want %v", err, errBoom)
		}
	})

	t.Run("sign error", func(t *testing.T) {
		hooks := Hooks{nil, okExtra, func(context.Context, string, time.Duration) error { return nil }}

		if _, err := hooks.RefreshAuthToken(context.Background(), testSecret, "HS256", "token-1", time.Hour); err == nil {
			t.Error("expected an error")
		}
	})

	t.Run("success", func(t *testing.T) {
		var gotID string
		var gotTTL time.Duration
		hooks := Hooks{okBuild, okExtra, func(_ context.Context, id string, ttl time.Duration) error {
			gotID, gotTTL = id, ttl
			return nil
		}}

		signed, err := hooks.RefreshAuthToken(context.Background(), testSecret, "HS256", "token-1", time.Hour)
		if err != nil {
			t.Fatalf("RefreshAuthToken() error = %v", err)
		}
		if _, err := ParseToken(signed, testSecret); err != nil {
			t.Errorf("the refreshed token is invalid: %v", err)
		}
		if gotID != "token-1" || gotTTL != time.Hour {
			t.Errorf("extend called with %q, %v", gotID, gotTTL)
		}
	})
}

func TestClaimGetters(t *testing.T) {
	epoch := time.Unix(0, 0)

	tests := []struct {
		name       string
		token      *jwt.Token
		wantUser   string
		wantToken  string
		wantUpdate time.Time
	}{
		{"nil token", nil, "", "", epoch},
		{"nil claims", &jwt.Token{}, "", "", epoch},
		{"empty claims", &jwt.Token{Claims: jwt.MapClaims{}}, "", "", epoch},
		{"registered claims", &jwt.Token{Claims: jwt.RegisteredClaims{}}, "", "", epoch},
		{
			"wrong types",
			&jwt.Token{Claims: jwt.MapClaims{"sub": 1, "id": true, "up": "now"}},
			"", "", epoch,
		},
		{
			"int64 update",
			&jwt.Token{Claims: jwt.MapClaims{"sub": "u", "id": "t", "up": int64(1700000000)}},
			"u", "t", time.Unix(1700000000, 0),
		},
		{
			"float64 update",
			&jwt.Token{Claims: jwt.MapClaims{"sub": "u", "id": "t", "up": float64(1700000000)}},
			"u", "t", time.Unix(1700000000, 0),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetUserID(tt.token); got != tt.wantUser {
				t.Errorf("GetUserID() = %q, want %q", got, tt.wantUser)
			}
			if got := GetTokenID(tt.token); got != tt.wantToken {
				t.Errorf("GetTokenID() = %q, want %q", got, tt.wantToken)
			}
			if got := GetLastUpdate(tt.token); !got.Equal(tt.wantUpdate) {
				t.Errorf("GetLastUpdate() = %v, want %v", got, tt.wantUpdate)
			}
		})
	}
}
