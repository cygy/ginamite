# Payments with Stripe

The plugin `plugins/stripe` takes payments with [Stripe](https://stripe.com). It creates the checkout sessions and the
payment intents, saves the payments, and handles the events of the webhook, so that the application delivers what was
paid for **once per payment**, even when Stripe sends an event several times.

| The plugin handles | The application handles |
| --- | --- |
| The calls to the API of Stripe, the amounts in cents, the metadata. | What is sold, its price, who may buy it. |
| The webhook: the signature, the live and test modes, the events sent again. | What a payment delivers, see `SucceededFunc`. |
| The payments saved, through a `Store` (`plugins/stripe/mongostore` for MongoDB). | The routes creating the payments, and the pages of Stripe.js. |

## Setup

```go
import (
	"github.com/cygy/ginamite/plugins/stripe"
	"github.com/cygy/ginamite/plugins/stripe/mongostore"
)

payments := stripe.NewPlugin(mongostore.New(appDatabase.Database, ""), deliver)
server.AddPlugin(payments)

// deliver is called once per payment, when Stripe says it succeeded.
func deliver(c *gin.Context, p stripe.Succeeded) error {
	return credit(c.Request.Context(), p.UserID, p.Metadata["packID"])
}
```

The settings are read from the section `stripe` of the configuration of the server:

```yaml
stripe:
  enabled: true
  secret_api_key: !env STRIPE_SECRET_KEY
  secret_webhook_key: !env STRIPE_WEBHOOK_SECRET
  publishable_api_key: pk_test_...     # for the pages, not used by the plugin
  currency: eur                         # eur by default
  payment_method_types: [card]          # chosen by Stripe from the account if empty
  webhook_path: /stripe/hook            # relative to the version of the API, /v1/stripe/hook by default
```

`Plugin.SettingsFunc` reads them from somewhere else instead, such as the configuration of the application. Without
`enabled`, the plugin adds no route and `Plugin.Service` returns nil.

**Live and test modes.** The payments are real ones only in the final version of the production
(`-env production -version final`): the webhook rejects the events of the test mode there, and the real ones
everywhere else. Use the test keys of Stripe outside the production.

## Creating a payment

The routes of the application call the service, once the plugin is configured:

```go
service := payments.Service()

checkout, err := service.CreateCheckout(ctx, stripe.CheckoutRequest{
	Customer:   stripe.Customer{UserID: userID, Email: email, StripeID: savedCustomerID},
	Items:      []stripe.Item{{Name: pack.Name, Amount: stripe.AmountInCents(pack.Price), Quantity: 1}},
	SuccessURL: successURL,
	CancelURL:  cancelURL,
	Metadata:   map[string]string{"packID": pack.ID},
	Data:       pack, // saved with the payment by the store
})
// Redirect to checkout.URL, or give checkout.SessionID to Stripe.js.
```

`CreateIntent` creates a payment intent instead, paid with Stripe.js on a page of the application: it returns the
client secret to give to Stripe.js.

Each payment has a **reference** chosen by the plugin, sent to Stripe in the metadata of the payment intent
(`paymentRef`, with the user in `userID`): since the version `2022-08-01` of the API of Stripe, a checkout session
creates its payment intent only when it is paid, so the payment can not be found by the ID of its payment intent. The
payment is saved with the status `todo`.

## The webhook

Declare the URL of the webhook (`https://api.example.com/v1/stripe/hook`) in the dashboard of Stripe, with the event
`payment_intent.succeeded`, and put its signing secret in `secret_webhook_key`. In development, the Stripe CLI forwards
the events: `stripe listen --forward-to http://localhost:8081/v1/stripe/hook`.

When a payment intent succeeds, the webhook:

1. checks the signature, and the mode of the event (live or test);
2. claims the payment in the store: only the first event of a payment goes on, the others are answered 200;
3. calls the `SucceededFunc` with the context of the request, which holds what the middlewares of the API saved
   (the database of the application, the queue). If it fails, the payment is released and the webhook answers 500: Stripe sends the
   event again later, so the function must deliver nothing, or be safe to call again, when it fails;
4. completes the payment with the IDs of the payment intent and of the customer of Stripe.

`Succeeded.CustomerID` is the customer Stripe created at the first checkout: save it with the user, and give it back in
`Customer.StripeID` so that the next checkouts reuse it.

The payment intents without reference, created before the plugin, are looked up with the ID of their payment intent.
The other events are answered 200 and ignored.

## Store

`Store` saves the payments: `Save`, `Claim` (in a single update, so that two events at once do not both win), `Release`
and `Complete`. `mongostore.Store` saves them in the collection `stripePayments` (or the one given to `mongostore.New`);
call `EnsureIndexes` at the start to create the unique index of the references. An application with its own collection
of payments implements `Store` on it, and finds its own data in `Payment.Data`.
