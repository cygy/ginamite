package stripe

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"
	stripeapi "github.com/stripe/stripe-go/v87"
)

// Customer is the user paying.
type Customer struct {
	// UserID is the ID of the user in the application, sent to Stripe in the metadata, see MetadataUserID.
	UserID string
	// Email receives the receipt. It also creates the customer of Stripe at the first checkout, if StripeID is empty.
	Email string
	// StripeID is the ID of the customer of Stripe, saved by the application after a first payment, see
	// Succeeded.CustomerID.
	StripeID string
}

// Item is a line of a checkout session.
type Item struct {
	Name string
	// Amount is the price of one item, in the smallest unit of the currency (cents), see AmountInCents.
	Amount   int64
	Quantity int64
}

// CheckoutRequest describes a checkout session to create.
type CheckoutRequest struct {
	Customer Customer
	Items    []Item
	// SuccessURL and CancelURL are the pages where Stripe sends the user back.
	SuccessURL string
	CancelURL  string
	// RequireBillingAddress makes Stripe ask for the billing address.
	RequireBillingAddress bool
	// Metadata are sent to Stripe with the payment intent, and given back to the SucceededFunc.
	Metadata map[string]string
	// Data is saved with the payment, see Payment.Data.
	Data any
}

// Checkout is a checkout session created.
type Checkout struct {
	// SessionID is the ID of the session, which Stripe.js redirects to.
	SessionID string
	// URL is the page of the session, to redirect to without Stripe.js.
	URL     string
	Payment *Payment
}

// IntentRequest describes a payment intent to create, paid with Stripe.js on a page of the application.
type IntentRequest struct {
	Customer Customer
	// Amount is the amount to pay, in the smallest unit of the currency (cents), see AmountInCents.
	Amount int64
	// Metadata are sent to Stripe with the payment intent, and given back to the SucceededFunc.
	Metadata map[string]string
	// Data is saved with the payment, see Payment.Data.
	Data any
}

// Intent is a payment intent created.
type Intent struct {
	// ClientSecret is given to Stripe.js to pay.
	ClientSecret string
	Payment      *Payment
}

// Service creates the payments with the API of Stripe, saves them in its Store, and handles the events of the webhook,
// see Service.Webhook.
type Service struct {
	settings  Settings
	client    *stripeapi.Client
	store     Store
	succeeded SucceededFunc
	// live is true if the service accepts only the real payments, false if it accepts only the test ones.
	live bool
	now  func() time.Time
}

// NewService returns a Service. live tells whether the payments are real ones, which must be true only in the final
// version of the production: the events of the other mode are rejected by the webhook. The options of the client are
// meant for the tests, to call a fake API.
func NewService(settings Settings, store Store, succeeded SucceededFunc, live bool, options ...stripeapi.ClientOption) *Service {
	settings = settings.withDefaults()

	return &Service{
		settings:  settings,
		client:    stripeapi.NewClient(settings.SecretAPIKey, options...),
		store:     store,
		succeeded: succeeded,
		live:      live,
		now:       time.Now,
	}
}

// Settings returns the settings of the service, with their default values.
func (s *Service) Settings() Settings {
	return s.settings
}

// AmountInCents returns a price in the unit of the currency (euros) as an amount in its smallest unit (cents). The price
// is rounded, not truncated: 19.99 * 100 is 1998.9999999999998 in floating point.
func AmountInCents(price float64) int64 {
	return int64(math.Round(price * 100))
}

// CreateCheckout creates a checkout session of Stripe and saves its payment, whose status is StatusTodo until Stripe
// says it is paid.
func (s *Service) CreateCheckout(ctx context.Context, request CheckoutRequest) (Checkout, error) {
	if len(request.Items) == 0 {
		return Checkout{}, errors.New("a checkout needs at least one item")
	}

	payment := s.newPayment(request.Customer, request.Metadata, request.Data)

	params := &stripeapi.CheckoutSessionCreateParams{
		Mode:              stripeapi.String(string(stripeapi.CheckoutSessionModePayment)),
		SuccessURL:        stripeapi.String(request.SuccessURL),
		CancelURL:         stripeapi.String(request.CancelURL),
		ClientReferenceID: stripeapi.String(payment.Reference),
		PaymentIntentData: &stripeapi.CheckoutSessionCreatePaymentIntentDataParams{
			Metadata: s.metadata(payment),
		},
	}
	if len(s.settings.PaymentMethodTypes) > 0 {
		params.AllowedPaymentMethodTypes = stripeapi.StringSlice(s.settings.PaymentMethodTypes)
	}
	if request.RequireBillingAddress {
		params.BillingAddressCollection = stripeapi.String(string(stripeapi.CheckoutSessionBillingAddressCollectionRequired))
	}
	if request.Customer.Email != "" {
		params.PaymentIntentData.ReceiptEmail = stripeapi.String(request.Customer.Email)
	}
	if request.Customer.StripeID != "" {
		params.Customer = stripeapi.String(request.Customer.StripeID)
	} else if request.Customer.Email != "" {
		params.CustomerEmail = stripeapi.String(request.Customer.Email)
	}

	for _, item := range request.Items {
		payment.Amount += item.Amount * item.Quantity
		params.LineItems = append(params.LineItems, &stripeapi.CheckoutSessionCreateLineItemParams{
			PriceData: &stripeapi.CheckoutSessionCreateLineItemPriceDataParams{
				Currency:   stripeapi.String(s.settings.Currency),
				UnitAmount: stripeapi.Int64(item.Amount),
				ProductData: &stripeapi.CheckoutSessionCreateLineItemPriceDataProductDataParams{
					Name: stripeapi.String(item.Name),
				},
			},
			Quantity: stripeapi.Int64(item.Quantity),
		})
	}

	session, err := s.client.V1CheckoutSessions.Create(ctx, params)
	if err != nil {
		return Checkout{}, fmt.Errorf("unable to create the checkout session: %w", err)
	}

	payment.CheckoutSessionID = session.ID
	payment.LiveMode = session.Livemode
	if err := s.store.Save(ctx, payment); err != nil {
		return Checkout{}, fmt.Errorf("unable to save the payment: %w", err)
	}

	return Checkout{SessionID: session.ID, URL: session.URL, Payment: payment}, nil
}

// CreateIntent creates a payment intent of Stripe and saves its payment, whose status is StatusTodo until Stripe says
// it is paid.
func (s *Service) CreateIntent(ctx context.Context, request IntentRequest) (Intent, error) {
	if request.Amount <= 0 {
		return Intent{}, errors.New("the amount of a payment intent must be positive")
	}

	payment := s.newPayment(request.Customer, request.Metadata, request.Data)
	payment.Amount = request.Amount

	params := &stripeapi.PaymentIntentCreateParams{
		Amount:   stripeapi.Int64(request.Amount),
		Currency: stripeapi.String(s.settings.Currency),
		Metadata: s.metadata(payment),
	}
	if request.Customer.Email != "" {
		params.ReceiptEmail = stripeapi.String(request.Customer.Email)
	}
	if request.Customer.StripeID != "" {
		params.Customer = stripeapi.String(request.Customer.StripeID)
	}

	intent, err := s.client.V1PaymentIntents.Create(ctx, params)
	if err != nil {
		return Intent{}, fmt.Errorf("unable to create the payment intent: %w", err)
	}

	payment.PaymentIntentID = intent.ID
	payment.LiveMode = intent.Livemode
	if err := s.store.Save(ctx, payment); err != nil {
		return Intent{}, fmt.Errorf("unable to save the payment: %w", err)
	}

	return Intent{ClientSecret: intent.ClientSecret, Payment: payment}, nil
}

// newPayment returns a new payment of the customer, with a new reference.
func (s *Service) newPayment(customer Customer, metadata map[string]string, data any) *Payment {
	now := s.now()

	return &Payment{
		Reference:  uuid.NewString(),
		UserID:     customer.UserID,
		CustomerID: customer.StripeID,
		Currency:   s.settings.Currency,
		Status:     StatusTodo,
		Metadata:   metadata,
		Data:       data,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}

// metadata returns the metadata of the payment intent: the ones of the application, then the reference and the user
// of the plugin, which win.
func (s *Service) metadata(payment *Payment) map[string]string {
	metadata := make(map[string]string, len(payment.Metadata)+2)
	for key, value := range payment.Metadata {
		metadata[key] = value
	}

	metadata[MetadataReference] = payment.Reference
	if payment.UserID != "" {
		metadata[MetadataUserID] = payment.UserID
	}

	return metadata
}
