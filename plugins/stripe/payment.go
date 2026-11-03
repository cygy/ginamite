package stripe

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
)

// Keys of the metadata the plugin sets on the payment intents. The application must not use them for its own metadata.
const (
	// MetadataReference is the key of the reference of the payment, see Payment.Reference.
	MetadataReference = "paymentRef"
	// MetadataUserID is the key of the ID of the user paying, see Customer.UserID.
	MetadataUserID = "userID"
)

// Status is the status of a payment.
type Status string

// The statuses of a payment.
const (
	// StatusTodo is the status of a payment created and not yet processed: it is not paid, or the event saying it is
	// paid has not been handled yet.
	StatusTodo Status = "todo"
	// StatusDone is the status of a payment processed: the application has delivered what was paid for.
	StatusDone Status = "done"
)

// Payment is a payment created by the Service, and saved by the Store until Stripe says it succeeded.
type Payment struct {
	// Reference identifies the payment in the Store. It is chosen by the Service and sent to Stripe in the metadata
	// of the payment intent, because the checkout sessions create their payment intent only when they are paid.
	Reference string
	// PaymentIntentID is the ID of the payment intent of Stripe. It is empty for a checkout session until it is paid,
	// see Store.Complete.
	PaymentIntentID string
	// CheckoutSessionID is the ID of the checkout session of Stripe, empty for a payment intent created directly.
	CheckoutSessionID string
	// UserID is the ID of the user paying.
	UserID string
	// CustomerID is the ID of the customer of Stripe, empty until Stripe knows it.
	CustomerID string
	// Amount is the amount to pay, in the smallest unit of the currency (cents).
	Amount   int64
	Currency string
	// LiveMode is true for a real payment, false for a payment of the test mode of Stripe.
	LiveMode bool
	Status   Status
	// Metadata are the metadata of the application sent to Stripe with the payment intent.
	Metadata map[string]string
	// Data is the application's own data about the payment, such as what is sold, saved by the Store as it is.
	Data      any
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Store saves the payments. Its methods must be safe for concurrent use: Stripe may deliver the same event several times
// at once.
type Store interface {
	// Save saves a new payment.
	Save(ctx context.Context, payment *Payment) error
	// Claim marks the payment with the reference as done if it was not yet, and reports whether it did. Only one
	// call returns true for a payment, so that an event delivered several times is handled once. It returns false,
	// and no error, if the payment is unknown.
	Claim(ctx context.Context, reference string) (bool, error)
	// Release marks the payment as not done again, after a failure of its handling, so that Stripe delivers it again.
	Release(ctx context.Context, reference string) error
	// Complete saves the IDs of the payment intent and of the customer of a payment handled.
	Complete(ctx context.Context, reference, paymentIntentID, customerID string) error
}

// Succeeded describes a payment Stripe says succeeded.
type Succeeded struct {
	Reference       string
	PaymentIntentID string
	// UserID is the ID of the user who paid, from the metadata of the payment intent.
	UserID string
	// CustomerID is the ID of the customer of Stripe, empty if Stripe did not create one.
	CustomerID string
	Amount     int64
	Currency   string
	LiveMode   bool
	// Metadata are the metadata of the payment intent, the ones of the application and the ones of the plugin.
	Metadata map[string]string
}

// SucceededFunc delivers what was paid for. It is called once per payment, after the payment is claimed in the Store,
// with the context of the request of the webhook, which holds what the middlewares of the API saved. If it returns an
// error, the payment is released and the webhook fails, so Stripe sends the event again later: it must deliver
// nothing, or be safe to call again, when it fails.
type SucceededFunc func(c *gin.Context, payment Succeeded) error
