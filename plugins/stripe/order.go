package stripe

import (
	"errors"

	"github.com/gin-gonic/gin"
)

// Order is what a user buys with a payment: the application describes it for a request of the routes creating the
// payments, see OrderFunc.
type Order struct {
	// Customer is the user paying. Its UserID is the authenticated user of the request if empty.
	Customer Customer
	// Items are the lines of the order: a payment intent is created for their total, a checkout session shows them. The
	// names are used by the checkout sessions only.
	Items []Item
	// RequireBillingAddress makes the checkout session ask for the billing address.
	RequireBillingAddress bool
	// Metadata are sent to Stripe with the payment intent, and given back to the SucceededFunc.
	Metadata map[string]string
	// Data is saved with the payment, see Payment.Data.
	Data any

	// SuccessRoute and CancelRoute are the keys of the web routes where Stripe sends the user back after a checkout,
	// in the file of the web routes of the server. Their URLs are built by the plugin in the locale of the request,
	// with SuccessParams and CancelParams replacing the parameters of the paths (":id", ":slug"...).
	SuccessRoute  string
	SuccessParams map[string]string
	CancelRoute   string
	CancelParams  map[string]string
}

// OrderFunc returns the order of a request of the routes creating the payments (intent or checkout), once the user is
// authenticated: it looks into the request for what is bought (the parameters of the path, the body...), checks that
// it can be bought and returns its price. It answers the request itself and returns false if the order is refused,
// such as an unknown or unavailable item.
type OrderFunc func(c *gin.Context) (Order, bool)

// total returns the total amount of the order, in the smallest unit of the currency.
func (o Order) total() int64 {
	var total int64
	for _, item := range o.Items {
		total += item.Amount * item.Quantity
	}

	return total
}

// validate returns an error if the order can not be paid.
func (o Order) validate() error {
	if len(o.Items) == 0 {
		return errors.New("the order has no item")
	}
	if o.total() <= 0 {
		return errors.New("the total of the order must be positive")
	}

	return nil
}
