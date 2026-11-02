package worker

import (
	"context"
	"testing"
	"time"

	"github.com/cygy/ginamite/common/notifications"
	"github.com/cygy/ginamite/worker/tasks/async/location"
)

// The fakes implement the storages with the function of the test. A method calls its function, which must be set.

type fakeTokens struct {
	deleteExpired func(ctx context.Context) (int64, error)
}

func (f fakeTokens) DeleteExpired(ctx context.Context) (int64, error) { return f.deleteExpired(ctx) }

type fakeAccounts struct {
	neverUsed func(ctx context.Context, days uint) ([]string, error)
	inactive  func(ctx context.Context, months uint) ([]string, error)
	isEmpty   func(ctx context.Context, userID string) (bool, error)
	sanitize  func(ctx context.Context) (uint, error)
}

func (f fakeAccounts) NeverUsed(ctx context.Context, days uint) ([]string, error) {
	return f.neverUsed(ctx, days)
}

func (f fakeAccounts) Inactive(ctx context.Context, months uint) ([]string, error) {
	return f.inactive(ctx, months)
}

func (f fakeAccounts) IsEmpty(ctx context.Context, userID string) (bool, error) {
	return f.isEmpty(ctx, userID)
}

func (f fakeAccounts) Sanitize(ctx context.Context) (uint, error) { return f.sanitize(ctx) }

type fakeUsers struct {
	del     func(ctx context.Context, userID string) (map[string]time.Time, error)
	disable func(ctx context.Context, userID string) (map[string]time.Time, error)
	enable  func(ctx context.Context, userID string) error
	social  func(ctx context.Context, userID string) error
}

func (f fakeUsers) Delete(ctx context.Context, userID string) (map[string]time.Time, error) {
	return f.del(ctx, userID)
}

func (f fakeUsers) Disable(ctx context.Context, userID string) (map[string]time.Time, error) {
	return f.disable(ctx, userID)
}

func (f fakeUsers) Enable(ctx context.Context, userID string) error { return f.enable(ctx, userID) }

func (f fakeUsers) UpdateSocialNetworks(ctx context.Context, userID string) error {
	return f.social(ctx, userID)
}

type fakeNotifications struct {
	targets func(ctx context.Context, userID, notificationType string) ([]notifications.Target, error)
}

func (f fakeNotifications) Targets(ctx context.Context, userID, notificationType string) ([]notifications.Target, error) {
	return f.targets(ctx, userID, notificationType)
}

type fakeLocations struct {
	save func(ctx context.Context, ipAddress, tokenID string, getDetails location.DetailsFunc) error
}

func (f fakeLocations) SaveIPAddressDetails(ctx context.Context, ipAddress, tokenID string, getDetails location.DetailsFunc) error {
	return f.save(ctx, ipAddress, tokenID, getDetails)
}

type fakeEmails struct {
	isValid func(ctx context.Context, emailAddress string) (bool, error)
}

func (f fakeEmails) IsValid(ctx context.Context, emailAddress string) (bool, error) {
	return f.isValid(ctx, emailAddress)
}

type fakeSitemap struct {
	variables func(ctx context.Context, routeKey string, offset, limit int) (map[string][]map[string]string, error)
}

func (f fakeSitemap) LocalizedRouteVariables(ctx context.Context, routeKey string, offset, limit int) (map[string][]map[string]string, error) {
	return f.variables(ctx, routeKey, offset, limit)
}

// allStorages returns storages which do nothing, except the ones of the tasks which log their count.
func allStorages() Storage {
	return Storage{
		Tokens: fakeTokens{deleteExpired: func(context.Context) (int64, error) { return 3, nil }},
		Accounts: fakeAccounts{
			neverUsed: func(context.Context, uint) ([]string, error) { return []string{"never-used"}, nil },
			inactive:  func(context.Context, uint) ([]string, error) { return []string{"inactive-1", "inactive-2"}, nil },
			isEmpty:   func(context.Context, string) (bool, error) { return true, nil },
			sanitize:  func(context.Context) (uint, error) { return 4, nil },
		},
		Users: fakeUsers{
			del:     func(context.Context, string) (map[string]time.Time, error) { return nil, nil },
			disable: func(context.Context, string) (map[string]time.Time, error) { return nil, nil },
			enable:  func(context.Context, string) error { return nil },
			social:  func(context.Context, string) error { return nil },
		},
		Notifications: fakeNotifications{targets: func(context.Context, string, string) ([]notifications.Target, error) { return nil, nil }},
		Locations:     fakeLocations{save: func(context.Context, string, string, location.DetailsFunc) error { return nil }},
		Emails:        fakeEmails{isValid: func(context.Context, string) (bool, error) { return true, nil }},
		Sitemap: fakeSitemap{variables: func(context.Context, string, int, int) (map[string][]map[string]string, error) {
			return nil, nil
		}},
	}
}

func TestStorage_undefined(t *testing.T) {
	var s Storage

	if s.targetsFunc() != nil || s.emailCheckFunc() != nil || s.deleteUserFunc() != nil || s.disableUserFunc() != nil ||
		s.enableUserFunc() != nil || s.updateSocialNetworksFunc() != nil || s.saveIPAddressDetailsFunc() != nil {
		t.Error("an undefined storage must give no function")
	}

	s = allStorages()
	if s.targetsFunc() == nil || s.emailCheckFunc() == nil || s.deleteUserFunc() == nil || s.disableUserFunc() == nil ||
		s.enableUserFunc() == nil || s.updateSocialNetworksFunc() == nil || s.saveIPAddressDetailsFunc() == nil {
		t.Error("a defined storage must give its functions")
	}
}
