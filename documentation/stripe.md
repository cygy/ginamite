# Payments with Stripe

The plugin `plugins/stripe` takes payments with [Stripe](https://stripe.com). It creates the checkout sessions and the
payment intents, saves the payments, and handles the events of the webhook, so that the application delivers what was
paid for **once per payment**, even when Stripe sends an event several times.

| The plugin handles | The application handles |
| --- | --- |
| The calls to the API of Stripe, the amounts in cents, the metadata. | What is sold, its price, who may buy it, see `OrderFunc`. |
| The routes creating the payments (intent, checkout), behind the authentication. | What a payment delivers, see `SucceededFunc`. |
| The webhook: the signature, the live and test modes, the events sent again. | The pages of the website, their routes, their templates and the JavaScript of Stripe.js. |
| The URLs where Stripe sends the user back, built from the web routes of the server. | |
| The payments saved, through a `Store` (`plugins/stripe/mongostore` for MongoDB). | |

## Setup

```go
import (
	"github.com/cygy/ginamite/plugins/stripe"
	"github.com/cygy/ginamite/plugins/stripe/mongostore"
)

payments := stripe.NewPlugin(mongostore.New(appDatabase.Database, ""), deliver)
payments.OrderFunc = order // required when the intents or the checkouts are enabled
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
  publishable_api_key: pk_test_...     # for the pages, see WebVars
  intent_enabled: true                  # adds the route creating the payment intents, false by default
  checkout_enabled: true                # adds the route creating the checkout sessions, false by default
  currency: eur                         # eur by default
  payment_method_types: [card]          # chosen by Stripe from the account if empty
  webhook_path: /stripe/hook            # relative to the version of the API, /v1/stripe/hook by default
  intent_path: /stripe/intent           # /v1/stripe/intent by default, may hold parameters (/store/packs/:id/intent)
  checkout_path: /stripe/checkout       # /v1/stripe/checkout by default
```

`Plugin.SettingsFunc` reads them from somewhere else instead, such as the configuration of the application. Without
`enabled`, the plugin adds no route and `Plugin.Service` returns nil.

**Live and test modes.** The payments are real ones only in the final version of the production
(`-env production -version final`): the webhook rejects the events of the test mode there, and the real ones
everywhere else. Use the test keys of Stripe outside the production.

## The routes creating the payments

When `intent_enabled` and `checkout_enabled` are set, the plugin adds `POST intent_path` and `POST checkout_path` to the
API. Both run, in this order: the middlewares of the API checking that the user is authenticated, that the
authentication is valid and that the user accepted the latest terms; the `OrderFunc` of the application; then the
creation of the payment. They answer `{"token": ...}`: the client secret of the payment intent, or the ID of the checkout
session, which Stripe.js uses.

The `OrderFunc` says what is bought. It reads the request (the parameters of the path, the body), checks that the user may
buy it, and answers the request itself if it refuses it, then returns `false`:

```go
func order(c *gin.Context) (stripe.Order, bool) {
	pack, err := findPack(c, c.Param("id"))
	if err != nil || pack.Price == 0 {
		response.PreconditionFailed(c) // the application chooses the answer
		return stripe.Order{}, false
	}

	return stripe.Order{
		Customer:              stripe.Customer{Email: user.Email, StripeID: user.StripeCustomerID},
		Items:                 []stripe.Item{{Name: pack.Name(ctxLocale.Get(c)), Amount: stripe.AmountInCents(pack.Price), Quantity: 1}},
		RequireBillingAddress: true,
		Metadata:              map[string]string{"packID": pack.ID},
		Data:                  pack,
		SuccessRoute:          "store.checkout.success", // keys of the file of the web routes
		CancelRoute:           "store",
	}, true
}
```

- `Customer.UserID` is the authenticated user if left empty.
- A payment intent is created for the total of the items, a checkout session shows them. An order without item, or whose
  total is not positive, is an error of the application: the route answers 500.
- The success and cancel URLs are built by the plugin with the web routes of the server (`-web-routes`, see
  `weburl.Of`) in the locale of the request. `SuccessParams` and `CancelParams` replace the parameters of their paths
  (`:slug`). Without web routes file they are empty and Stripe refuses the checkout.
- When the Stripe call or the store fails, the route answers 500 and logs the error.

The handlers are also available as `Service.IntentHandler(orderFunc)` and `Service.CheckoutHandler(orderFunc)`, to add
to routes of your own with other middlewares.

## The pages

The plugin does not render the pages: their localized routes (`routes.yml`), their templates and the JavaScript of
Stripe.js belong to the application, as a plugin has no way to add pages to the web server. `Plugin.WebVars()` gives the
variables they need, read from the settings: `StripePublishableKey`, `IntentEnabled` and `CheckoutEnabled` (the
constants `VarPublishableKey`, `VarIntentEnabled` and `VarCheckoutEnabled`):

```go
for name, value := range payments.WebVars() {
	page.SetVar(name, value)
}
```

The web server runs the plugin too, with its own configuration file: its section `stripe` only needs `enabled`,
`publishable_api_key`, `intent_enabled` and `checkout_enabled`. The secrets are not required, and the store and the
`OrderFunc` can be nil there (`Server.Role()` tells the plugin which server runs).

## Creating a payment

The routes above use the service. The routes of the application can call it directly, once the plugin is configured:

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
