package stripe

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/cygy/ginamite/common/log"

	"github.com/gin-gonic/gin"
	stripeapi "github.com/stripe/stripe-go/v87"
	"github.com/stripe/stripe-go/v87/webhook"
)

// maxWebhookBodySize is the size limit of the body of an event, the one Stripe recommends.
const maxWebhookBodySize = 65536

// eventPaymentIntentSucceeded is the type of the event sent when a payment intent is paid.
const eventPaymentIntentSucceeded = "payment_intent.succeeded"

// Webhook handles the events Stripe sends, whose signature it checks. When a payment intent succeeds, it claims its
// payment in the Store, calls the SucceededFunc, then completes the payment. It answers:
//   - 200 to the events handled, to the other types of events, and to the payments unknown or already handled;
//   - 400 to the events with an invalid signature, which can not be read, or from the other mode (live or test);
//   - 500 if the Store or the SucceededFunc fails, so that Stripe sends the event again later.
func (s *Service) Webhook(c *gin.Context) {
	payload, err := io.ReadAll(http.MaxBytesReader(c.Writer, c.Request.Body, maxWebhookBodySize))
	if err != nil {
		log.Error("unable to read the stripe event", "error", err.Error())
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	// The version of the API of the event is the one of the Stripe account, which may differ from the one of the
	// library: only the stable fields of the payment intent are read.
	event, err := webhook.ConstructEventWithOptions(payload, c.GetHeader("Stripe-Signature"), s.settings.SecretWebhookKey,
		webhook.ConstructEventOptions{IgnoreAPIVersionMismatch: true})
	if err != nil {
		log.Error("invalid stripe event", "error", err.Error())
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	if event.Type != eventPaymentIntentSucceeded {
		c.Status(http.StatusOK)
		return
	}

	var intent stripeapi.PaymentIntent
	if err := json.Unmarshal(event.Data.Raw, &intent); err != nil {
		log.Error("unable to read the payment intent of the stripe event", "event", event.ID, "error", err.Error())
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	c.Status(s.handleSucceeded(c, &intent))
}

// handleSucceeded handles a payment intent paid, and returns the status of the response.
func (s *Service) handleSucceeded(c *gin.Context, intent *stripeapi.PaymentIntent) int {
	if intent.Livemode != s.live {
		log.Error("stripe event of the wrong mode", "paymentIntent", intent.ID, "live", intent.Livemode)
		return http.StatusBadRequest
	}

	ctx := c.Request.Context()

	// The payments created before the plugin have no reference: they are saved with the ID of their payment intent.
	reference := intent.Metadata[MetadataReference]
	if reference == "" {
		reference = intent.ID
	}

	claimed, err := s.store.Claim(ctx, reference)
	if err != nil {
		log.Error("unable to claim the payment", "reference", reference, "error", err.Error())
		return http.StatusInternalServerError
	}
	if !claimed {
		log.Warn("payment unknown or already handled", "reference", reference, "paymentIntent", intent.ID)
		return http.StatusOK
	}

	var customerID string
	if intent.Customer != nil {
		customerID = intent.Customer.ID
	}

	succeeded := Succeeded{
		Reference:       reference,
		PaymentIntentID: intent.ID,
		UserID:          intent.Metadata[MetadataUserID],
		CustomerID:      customerID,
		Amount:          intent.Amount,
		Currency:        string(intent.Currency),
		LiveMode:        intent.Livemode,
		Metadata:        intent.Metadata,
	}
	if err := s.succeeded(c, succeeded); err != nil {
		log.Error("unable to deliver the payment", "reference", reference, "error", err.Error())
		if err := s.store.Release(ctx, reference); err != nil {
			log.Error("unable to release the payment", "reference", reference, "error", err.Error())
		}
		return http.StatusInternalServerError
	}

	// The payment is delivered: a failure is only logged, Stripe must not send the event again.
	if err := s.store.Complete(ctx, reference, intent.ID, customerID); err != nil {
		log.Error("unable to complete the payment", "reference", reference, "error", err.Error())
	}

	return http.StatusOK
}
