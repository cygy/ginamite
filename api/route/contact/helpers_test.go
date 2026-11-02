package contact

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"github.com/cygy/ginamite/common/ctxkey"

	ctxLocale "github.com/cygy/ginamite/api/context/locale"
	"github.com/cygy/ginamite/api/internal/apitest"

	"github.com/gin-gonic/gin"
)

// fakeStorage is a Storage whose methods call the function set by the test, if any.
type fakeStorage struct {
	create func(ctx context.Context, emailAddress, subject, text string) error
	update func(ctx context.Context, contactID string, done bool) error
	list   func(ctx context.Context, offset, limit int64) (any, error)
	get    func(ctx context.Context, contactID string) (any, error)
}

func (f *fakeStorage) CreateContact(ctx context.Context, emailAddress, subject, text string) error {
	if f.create == nil {
		return nil
	}

	return f.create(ctx, emailAddress, subject, text)
}

func (f *fakeStorage) UpdateContact(ctx context.Context, contactID string, done bool) error {
	if f.update == nil {
		return nil
	}

	return f.update(ctx, contactID, done)
}

func (f *fakeStorage) GetContacts(ctx context.Context, offset, limit int64) (any, error) {
	if f.list == nil {
		return nil, nil
	}

	return f.list(ctx, offset, limit)
}

func (f *fakeStorage) GetContact(ctx context.Context, contactID string) (any, error) {
	if f.get == nil {
		return nil, nil
	}

	return f.get(ctx, contactID)
}

// testStorage is the storage of the handlers, reset at the end of each test.
var testStorage = &fakeStorage{}

func setup(t *testing.T) {
	t.Helper()

	apitest.InitLocalization(t, map[string]string{"placeholder": "placeholder"})
	t.Cleanup(func() { *testStorage = fakeStorage{} })
}

func serve(handler gin.HandlerFunc, method, target, body string, middlewares ...gin.HandlerFunc) int {
	router := gin.New()
	router.Use(Inject(testStorage), ctxkey.Localization.Inject(apitest.Localization), ctxLocale.Inject("en", []string{"en"}))
	router.Use(middlewares...)
	router.Any("/:id", handler)
	router.Any("/", handler)

	req, _ := http.NewRequest(method, target, strings.NewReader(body))

	return apitest.Serve(router, req).Code
}

// withoutStorage is a middleware removing the storage, as an application which provides none.
func withoutStorage(c *gin.Context) { SetStorage(c, nil) }
