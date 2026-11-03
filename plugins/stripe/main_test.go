package stripe

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"sync"
	"testing"

	"github.com/cygy/ginamite/common/log"
	"github.com/cygy/ginamite/common/log/logtest"

	stripeapi "github.com/stripe/stripe-go/v87"
)

func TestMain(m *testing.M) {
	log.Main, _ = logtest.New()

	os.Exit(m.Run())
}

// fakeStore is a Store in memory, whose methods may fail.
type fakeStore struct {
	mu       sync.Mutex
	payments map[string]*Payment
	saved    []*Payment
	released []string
	complete []string

	saveErr, claimErr, releaseErr, completeErr error
}

func newFakeStore() *fakeStore {
	return &fakeStore{payments: map[string]*Payment{}}
}

func (f *fakeStore) Save(_ context.Context, payment *Payment) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.saveErr != nil {
		return f.saveErr
	}
	f.payments[payment.Reference] = payment
	f.saved = append(f.saved, payment)

	return nil
}

func (f *fakeStore) Claim(_ context.Context, reference string) (bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.claimErr != nil {
		return false, f.claimErr
	}
	payment, ok := f.payments[reference]
	if !ok || payment.Status == StatusDone {
		return false, nil
	}
	payment.Status = StatusDone

	return true, nil
}

func (f *fakeStore) Release(_ context.Context, reference string) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.released = append(f.released, reference)
	if payment, ok := f.payments[reference]; ok {
		payment.Status = StatusTodo
	}

	return f.releaseErr
}

func (f *fakeStore) Complete(_ context.Context, reference, paymentIntentID, customerID string) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.complete = append(f.complete, reference+"/"+paymentIntentID+"/"+customerID)

	return f.completeErr
}

// fakeAPI is a fake API of Stripe, answering the creation of the checkout sessions and of the payment intents.
type fakeAPI struct {
	mu     sync.Mutex
	server *httptest.Server
	forms  map[string]url.Values
	fail   bool
}

func newFakeAPI(t *testing.T) *fakeAPI {
	t.Helper()

	api := &fakeAPI{forms: map[string]url.Values{}}
	api.server = httptest.NewServer(http.HandlerFunc(api.serve))
	t.Cleanup(api.server.Close)

	return api
}

func (a *fakeAPI) serve(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	a.mu.Lock()
	a.forms[r.URL.Path] = r.PostForm
	fail := a.fail
	a.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	if fail {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]any{"error": map[string]string{"type": "invalid_request_error", "message": "refused"}})
		return
	}

	switch r.URL.Path {
	case "/v1/checkout/sessions":
		_ = json.NewEncoder(w).Encode(map[string]any{"id": "cs_test_1", "object": "checkout.session", "url": "https://checkout.stripe.com/c/cs_test_1", "livemode": false})
	case "/v1/payment_intents":
		_ = json.NewEncoder(w).Encode(map[string]any{"id": "pi_test_1", "object": "payment_intent", "client_secret": "pi_test_1_secret", "livemode": false})
	default:
		http.NotFound(w, r)
	}
}

// form returns the form sent to the path.
func (a *fakeAPI) form(path string) url.Values {
	a.mu.Lock()
	defer a.mu.Unlock()

	return a.forms[path]
}

// options returns the options of a client calling the fake API.
func (a *fakeAPI) options() []stripeapi.ClientOption {
	backends := stripeapi.NewBackendsWithConfig(&stripeapi.BackendConfig{
		URL:               stripeapi.String(a.server.URL),
		HTTPClient:        a.server.Client(),
		MaxNetworkRetries: stripeapi.Int64(0),
		LeveledLogger:     &stripeapi.LeveledLogger{Level: stripeapi.LevelNull},
	})

	return []stripeapi.ClientOption{stripeapi.WithBackends(backends)}
}

// testSettings are valid settings.
func testSettings() Settings {
	return Settings{Enabled: true, SecretAPIKey: "sk_test_key", SecretWebhookKey: "whsec_test"}
}

var errTest = errors.New("test failure")
