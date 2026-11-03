package stripe

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stripe/stripe-go/v87/webhook"
)

// event returns the JSON of an event of the type, whose object is the payment intent.
func event(t *testing.T, eventType string, intent map[string]any) []byte {
	t.Helper()

	payload, err := json.Marshal(map[string]any{
		"id":          "evt_1",
		"object":      "event",
		"type":        eventType,
		"api_version": "2020-08-27",
		"data":        map[string]any{"object": intent},
	})
	if err != nil {
		t.Fatal(err)
	}

	return payload
}

// intent returns a payment intent paid, with the metadata.
func intent(live bool, metadata map[string]string) map[string]any {
	return map[string]any{
		"id":       "pi_1",
		"object":   "payment_intent",
		"amount":   499,
		"currency": "eur",
		"livemode": live,
		"customer": "cus_1",
		"metadata": metadata,
	}
}

// post sends the payload, signed with the secret, to the webhook of the service, and returns the status.
func post(t *testing.T, s *Service, payload []byte, secret string) int {
	t.Helper()

	signed := webhook.GenerateTestSignedPayload(&webhook.UnsignedPayload{Payload: payload, Secret: secret})
	request := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/stripe/hook", bytes.NewReader(payload))
	request.Header.Set("Stripe-Signature", signed.Header)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/stripe/hook", s.Webhook)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)

	return recorder.Code
}

// webhookService returns a service in test mode whose store holds a payment with the reference "ref1".
func webhookService(store *fakeStore, succeeded SucceededFunc) *Service {
	store.payments["ref1"] = &Payment{Reference: "ref1", Status: StatusTodo}
	if succeeded == nil {
		succeeded = func(*gin.Context, Succeeded) error { return nil }
	}

	return NewService(testSettings(), store, succeeded, false)
}

func TestService_Webhook(t *testing.T) {
	metadata := map[string]string{MetadataReference: "ref1", MetadataUserID: "user1", "packID": "pack1"}

	t.Run("payment delivered once", func(t *testing.T) {
		store := newFakeStore()
		var delivered []Succeeded
		s := webhookService(store, func(_ *gin.Context, p Succeeded) error {
			delivered = append(delivered, p)
			return nil
		})
		payload := event(t, eventPaymentIntentSucceeded, intent(false, metadata))

		if status := post(t, s, payload, "whsec_test"); status != http.StatusOK {
			t.Fatalf("status = %d", status)
		}
		if status := post(t, s, payload, "whsec_test"); status != http.StatusOK {
			t.Fatalf("status of the event sent again = %d", status)
		}

		if len(delivered) != 1 {
			t.Fatalf("delivered %d times, want once", len(delivered))
		}
		p := delivered[0]
		if p.Reference != "ref1" || p.PaymentIntentID != "pi_1" || p.UserID != "user1" || p.CustomerID != "cus_1" ||
			p.Amount != 499 || p.Currency != "eur" || p.LiveMode || p.Metadata["packID"] != "pack1" {
			t.Errorf("succeeded = %+v", p)
		}
		if len(store.complete) != 1 || store.complete[0] != "ref1/pi_1/cus_1" {
			t.Errorf("completed = %v", store.complete)
		}
	})

	t.Run("payment created before the plugin", func(t *testing.T) {
		store := newFakeStore()
		s := webhookService(store, nil)
		store.payments["pi_1"] = &Payment{Reference: "pi_1", Status: StatusTodo}

		if status := post(t, s, event(t, eventPaymentIntentSucceeded, intent(false, nil)), "whsec_test"); status != http.StatusOK {
			t.Fatalf("status = %d", status)
		}
		if store.payments["pi_1"].Status != StatusDone {
			t.Error("the payment saved with the ID of its payment intent must be claimed")
		}
	})

	t.Run("delivery failure", func(t *testing.T) {
		store := newFakeStore()
		s := webhookService(store, func(*gin.Context, Succeeded) error { return errTest })
		store.releaseErr = errTest

		if status := post(t, s, event(t, eventPaymentIntentSucceeded, intent(false, metadata)), "whsec_test"); status != http.StatusInternalServerError {
			t.Errorf("status = %d, want 500", status)
		}
		if len(store.released) != 1 || len(store.complete) != 0 {
			t.Errorf("released = %v, completed = %v", store.released, store.complete)
		}
	})

	t.Run("completion failure", func(t *testing.T) {
		store := newFakeStore()
		store.completeErr = errTest

		if status := post(t, webhookService(store, nil), event(t, eventPaymentIntentSucceeded, intent(false, metadata)), "whsec_test"); status != http.StatusOK {
			t.Errorf("status = %d, want 200: the payment is delivered", status)
		}
	})

	t.Run("claim failure", func(t *testing.T) {
		store := newFakeStore()
		store.claimErr = errTest

		if status := post(t, webhookService(store, nil), event(t, eventPaymentIntentSucceeded, intent(false, metadata)), "whsec_test"); status != http.StatusInternalServerError {
			t.Errorf("status = %d, want 500", status)
		}
	})

	t.Run("unknown payment", func(t *testing.T) {
		unknown := map[string]string{MetadataReference: "other"}
		if status := post(t, webhookService(newFakeStore(), nil), event(t, eventPaymentIntentSucceeded, intent(false, unknown)), "whsec_test"); status != http.StatusOK {
			t.Errorf("status = %d, want 200", status)
		}
	})

	t.Run("rejected", func(t *testing.T) {
		tests := map[string]struct {
			payload []byte
			secret  string
		}{
			"invalid signature": {event(t, eventPaymentIntentSucceeded, intent(false, metadata)), "whsec_other"},
			"wrong mode":        {event(t, eventPaymentIntentSucceeded, intent(true, metadata)), "whsec_test"},
			"invalid object":    {event(t, eventPaymentIntentSucceeded, map[string]any{"id": "pi_1", "amount": "many"}), "whsec_test"},
			"too large":         {[]byte(`{"id":"` + strings.Repeat("a", maxWebhookBodySize) + `"}`), "whsec_test"},
		}

		for name, test := range tests {
			t.Run(name, func(t *testing.T) {
				store := newFakeStore()
				if status := post(t, webhookService(store, nil), test.payload, test.secret); status != http.StatusBadRequest {
					t.Errorf("status = %d, want 400", status)
				}
				if store.payments["ref1"].Status != StatusTodo {
					t.Error("the payment must not be claimed")
				}
			})
		}
	})

	t.Run("other event", func(t *testing.T) {
		store := newFakeStore()
		if status := post(t, webhookService(store, nil), event(t, "payment_intent.created", intent(false, metadata)), "whsec_test"); status != http.StatusOK {
			t.Errorf("status = %d, want 200", status)
		}
		if store.payments["ref1"].Status != StatusTodo {
			t.Error("the payment must not be claimed")
		}
	})
}
