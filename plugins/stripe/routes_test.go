package stripe

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	ctxLocale "github.com/cygy/ginamite/api/context/locale"
	"github.com/cygy/ginamite/common/weburl"

	"github.com/gin-gonic/gin"
)

// webRoutes returns web routes with the success and cancel pages of a checkout.
func webRoutes(t *testing.T) *weburl.Configuration {
	t.Helper()

	file := filepath.Join(t.TempDir(), "routes.yml")
	content := "routes:\n  store:\n    paths:\n      en-us: /{{.Locale}}/store\n  store.success:\n    paths:\n      en-us: /{{.Locale}}/store/:slug/success\n"
	if err := os.WriteFile(file, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	routes, err := weburl.Load(file, "https://example.com")
	if err != nil {
		t.Fatal(err)
	}

	return routes
}

// serve posts a request to the handler, in the locale en-us with the web routes, and returns the response.
func serve(t *testing.T, handler gin.HandlerFunc, routes *weburl.Configuration) *httptest.ResponseRecorder {
	t.Helper()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(weburl.Inject(routes), func(c *gin.Context) { ctxLocale.Set(c, "en-us", "") })
	router.POST("/pay/:id", handler)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/pay/pack1", nil))

	return recorder
}

// token returns the token of a response.
func token(t *testing.T, recorder *httptest.ResponseRecorder) string {
	t.Helper()

	var body struct{ Token string }
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("body = %q: %v", recorder.Body.String(), err)
	}

	return body.Token
}

// testOrder is a valid order of the pack in the path.
func testOrder(c *gin.Context) (Order, bool) {
	return Order{
		Customer:              Customer{UserID: "user1", Email: "player@example.com"},
		Items:                 []Item{{Name: "Pack " + c.Param("id"), Amount: 499, Quantity: 1}},
		RequireBillingAddress: true,
		Metadata:              map[string]string{"packID": c.Param("id")},
		Data:                  "pack data",
		SuccessRoute:          "store.success",
		SuccessParams:         map[string]string{"slug": "bought"},
		CancelRoute:           "store",
	}, true
}

func TestService_IntentHandler(t *testing.T) {
	t.Run("created", func(t *testing.T) {
		api, store := newFakeAPI(t), newFakeStore()
		recorder := serve(t, newTestService(api, store, nil, false).IntentHandler(testOrder), nil)

		if recorder.Code != http.StatusOK || token(t, recorder) != "pi_test_1_secret" {
			t.Fatalf("response = %d %s", recorder.Code, recorder.Body)
		}
		form := api.form("/v1/payment_intents")
		if form.Get("amount") != "499" || form.Get("metadata[packID]") != "pack1" || form.Get("receipt_email") != "player@example.com" {
			t.Errorf("form = %v", form)
		}
		if len(store.saved) != 1 || store.saved[0].Data != "pack data" || store.saved[0].UserID != "user1" {
			t.Errorf("saved = %+v", store.saved)
		}
	})

	t.Run("refused by the application", func(t *testing.T) {
		api := newFakeAPI(t)
		refuse := func(c *gin.Context) (Order, bool) {
			c.AbortWithStatus(http.StatusPreconditionFailed)
			return Order{}, false
		}

		recorder := serve(t, newTestService(api, newFakeStore(), nil, false).IntentHandler(refuse), nil)
		if recorder.Code != http.StatusPreconditionFailed || api.form("/v1/payment_intents") != nil {
			t.Errorf("code = %d, form = %v, want 412 and no call to Stripe", recorder.Code, api.form("/v1/payment_intents"))
		}
	})

	t.Run("invalid order", func(t *testing.T) {
		api := newFakeAPI(t)
		free := func(*gin.Context) (Order, bool) { return Order{Items: []Item{{Name: "Free", Quantity: 1}}}, true }

		recorder := serve(t, newTestService(api, newFakeStore(), nil, false).IntentHandler(free), nil)
		if recorder.Code != http.StatusInternalServerError || api.form("/v1/payment_intents") != nil {
			t.Errorf("code = %d, want 500 and no call to Stripe", recorder.Code)
		}
	})

	t.Run("API failure", func(t *testing.T) {
		api := newFakeAPI(t)
		api.fail = true

		if recorder := serve(t, newTestService(api, newFakeStore(), nil, false).IntentHandler(testOrder), nil); recorder.Code != http.StatusInternalServerError {
			t.Errorf("code = %d, want 500", recorder.Code)
		}
	})
}

func TestService_CheckoutHandler(t *testing.T) {
	t.Run("created with the URLs of the web routes", func(t *testing.T) {
		api, store := newFakeAPI(t), newFakeStore()
		recorder := serve(t, newTestService(api, store, nil, false).CheckoutHandler(testOrder), webRoutes(t))

		if recorder.Code != http.StatusOK || token(t, recorder) != "cs_test_1" {
			t.Fatalf("response = %d %s", recorder.Code, recorder.Body)
		}
		form := api.form("/v1/checkout/sessions")
		want := map[string]string{
			"success_url":                            "https://example.com/en-us/store/bought/success",
			"cancel_url":                             "https://example.com/en-us/store",
			"billing_address_collection":             "required",
			"line_items[0][price_data][unit_amount]": "499",
			"payment_intent_data[metadata][packID]":  "pack1",
		}
		for key, value := range want {
			if got := form.Get(key); got != value {
				t.Errorf("%s = %q, want %q", key, got, value)
			}
		}
		if len(store.saved) != 1 || store.saved[0].CheckoutSessionID != "cs_test_1" {
			t.Errorf("saved = %+v", store.saved)
		}
	})

	t.Run("refused by the application", func(t *testing.T) {
		refuse := func(c *gin.Context) (Order, bool) {
			c.AbortWithStatus(http.StatusPreconditionFailed)
			return Order{}, false
		}

		if recorder := serve(t, newTestService(newFakeAPI(t), newFakeStore(), nil, false).CheckoutHandler(refuse), nil); recorder.Code != http.StatusPreconditionFailed {
			t.Errorf("code = %d, want 412", recorder.Code)
		}
	})

	t.Run("API failure", func(t *testing.T) {
		api := newFakeAPI(t)
		api.fail = true

		if recorder := serve(t, newTestService(api, newFakeStore(), nil, false).CheckoutHandler(testOrder), webRoutes(t)); recorder.Code != http.StatusInternalServerError {
			t.Errorf("code = %d, want 500", recorder.Code)
		}
	})
}
