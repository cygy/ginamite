package stripe

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestAmountInCents(t *testing.T) {
	tests := map[float64]int64{0: 0, 1: 100, 19.99: 1999, 0.29: 29, 4.995: 500}
	for price, want := range tests {
		if got := AmountInCents(price); got != want {
			t.Errorf("AmountInCents(%v) = %d, want %d", price, got, want)
		}
	}
}

func TestNewService(t *testing.T) {
	s := NewService(Settings{SecretAPIKey: "sk"}, newFakeStore(), nil, true)

	if s.Settings().Currency != DefaultCurrency || s.Settings().WebhookPath != DefaultWebhookPath || !s.live {
		t.Errorf("service = %+v", s)
	}
}

// newTestService returns a service calling the fake API, with a fixed clock.
func newTestService(api *fakeAPI, store Store, succeeded SucceededFunc, live bool) *Service {
	settings := testSettings()
	settings.PaymentMethodTypes = []string{"card"}

	s := NewService(settings, store, succeeded, live, api.options()...)
	s.now = func() time.Time { return time.Date(2026, 11, 3, 10, 0, 0, 0, time.UTC) }

	return s
}

func TestService_CreateCheckout(t *testing.T) {
	request := CheckoutRequest{
		Customer:              Customer{UserID: "user1", Email: "player@example.com"},
		Items:                 []Item{{Name: "Pack", Amount: 499, Quantity: 2}},
		SuccessURL:            "https://example.com/success",
		CancelURL:             "https://example.com/store",
		RequireBillingAddress: true,
		Metadata:              map[string]string{"packID": "pack1", MetadataUserID: "spoofed"},
		Data:                  "pack data",
	}

	t.Run("created and saved", func(t *testing.T) {
		api, store := newFakeAPI(t), newFakeStore()

		checkout, err := newTestService(api, store, nil, false).CreateCheckout(context.Background(), request)
		if err != nil {
			t.Fatal(err)
		}

		payment := checkout.Payment
		if checkout.SessionID != "cs_test_1" || checkout.URL == "" || len(store.saved) != 1 || store.saved[0] != payment {
			t.Fatalf("checkout = %+v, saved = %v", checkout, store.saved)
		}
		if payment.Reference == "" || payment.CheckoutSessionID != "cs_test_1" || payment.PaymentIntentID != "" ||
			payment.UserID != "user1" || payment.Amount != 998 || payment.Currency != "eur" || payment.Status != StatusTodo ||
			payment.Data != "pack data" || payment.CreatedAt.IsZero() {
			t.Errorf("payment = %+v", payment)
		}

		form := api.form("/v1/checkout/sessions")
		want := map[string]string{
			"mode":                                      "payment",
			"client_reference_id":                       payment.Reference,
			"customer_email":                            "player@example.com",
			"billing_address_collection":                "required",
			"allowed_payment_method_types[0]":           "card",
			"payment_intent_data[receipt_email]":        "player@example.com",
			"payment_intent_data[metadata][paymentRef]": payment.Reference,
			"payment_intent_data[metadata][userID]":     "user1",
			"payment_intent_data[metadata][packID]":     "pack1",
			"line_items[0][quantity]":                   "2",
			"line_items[0][price_data][unit_amount]":    "499",
			"line_items[0][price_data][currency]":       "eur",
		}
		for key, value := range want {
			if got := form.Get(key); got != value {
				t.Errorf("%s = %q, want %q", key, got, value)
			}
		}
	})

	t.Run("known customer", func(t *testing.T) {
		api := newFakeAPI(t)
		r := request
		r.Customer.StripeID = "cus_1"

		checkout, err := newTestService(api, newFakeStore(), nil, false).CreateCheckout(context.Background(), r)
		if err != nil {
			t.Fatal(err)
		}

		form := api.form("/v1/checkout/sessions")
		if form.Get("customer") != "cus_1" || form.Get("customer_email") != "" || checkout.Payment.CustomerID != "cus_1" {
			t.Errorf("form = %v", form)
		}
	})

	t.Run("no item", func(t *testing.T) {
		if _, err := newTestService(newFakeAPI(t), newFakeStore(), nil, false).CreateCheckout(context.Background(), CheckoutRequest{}); err == nil {
			t.Error("CreateCheckout() = nil, want an error")
		}
	})

	t.Run("API failure", func(t *testing.T) {
		api, store := newFakeAPI(t), newFakeStore()
		api.fail = true

		if _, err := newTestService(api, store, nil, false).CreateCheckout(context.Background(), request); err == nil || len(store.saved) != 0 {
			t.Errorf("CreateCheckout() = %v, saved = %d, want an error and nothing saved", err, len(store.saved))
		}
	})

	t.Run("store failure", func(t *testing.T) {
		store := newFakeStore()
		store.saveErr = errTest

		if _, err := newTestService(newFakeAPI(t), store, nil, false).CreateCheckout(context.Background(), request); !errors.Is(err, errTest) {
			t.Errorf("CreateCheckout() = %v, want %v", err, errTest)
		}
	})
}

func TestService_CreateIntent(t *testing.T) {
	request := IntentRequest{
		Customer: Customer{UserID: "user1", Email: "player@example.com", StripeID: "cus_1"},
		Amount:   1999,
		Metadata: map[string]string{"packID": "pack1"},
	}

	t.Run("created and saved", func(t *testing.T) {
		api, store := newFakeAPI(t), newFakeStore()

		intent, err := newTestService(api, store, nil, false).CreateIntent(context.Background(), request)
		if err != nil {
			t.Fatal(err)
		}

		payment := intent.Payment
		if intent.ClientSecret != "pi_test_1_secret" || payment.PaymentIntentID != "pi_test_1" || payment.Amount != 1999 || len(store.saved) != 1 {
			t.Fatalf("intent = %+v, payment = %+v", intent, payment)
		}

		form := api.form("/v1/payment_intents")
		want := map[string]string{
			"amount":               "1999",
			"currency":             "eur",
			"customer":             "cus_1",
			"receipt_email":        "player@example.com",
			"metadata[paymentRef]": payment.Reference,
			"metadata[userID]":     "user1",
			"metadata[packID]":     "pack1",
		}
		for key, value := range want {
			if got := form.Get(key); got != value {
				t.Errorf("%s = %q, want %q", key, got, value)
			}
		}
	})

	t.Run("invalid amount", func(t *testing.T) {
		if _, err := newTestService(newFakeAPI(t), newFakeStore(), nil, false).CreateIntent(context.Background(), IntentRequest{}); err == nil {
			t.Error("CreateIntent() = nil, want an error")
		}
	})

	t.Run("API failure", func(t *testing.T) {
		api := newFakeAPI(t)
		api.fail = true

		if _, err := newTestService(api, newFakeStore(), nil, false).CreateIntent(context.Background(), request); err == nil {
			t.Error("CreateIntent() = nil, want an error")
		}
	})

	t.Run("store failure", func(t *testing.T) {
		store := newFakeStore()
		store.saveErr = errTest

		if _, err := newTestService(newFakeAPI(t), store, nil, false).CreateIntent(context.Background(), request); !errors.Is(err, errTest) {
			t.Errorf("CreateIntent() = %v, want %v", err, errTest)
		}
	})
}
