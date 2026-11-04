package stripe

import (
	ctxLocale "github.com/cygy/ginamite/api/context/locale"
	ctxUser "github.com/cygy/ginamite/api/context/user"
	"github.com/cygy/ginamite/api/response"
	"github.com/cygy/ginamite/common/log"
	"github.com/cygy/ginamite/common/weburl"

	"github.com/gin-gonic/gin"
)

// IntentHandler returns the handler of the route creating a payment intent, paid with Stripe.js. It asks the order to
// orderFunc, then answers {"token": <client secret>} for Stripe.js, or 500 if Stripe or the Store fails.
func (s *Service) IntentHandler(orderFunc OrderFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		order, ok := readOrder(c, orderFunc)
		if !ok {
			return
		}

		intent, err := s.CreateIntent(c.Request.Context(), IntentRequest{
			Customer: order.Customer,
			Amount:   order.total(),
			Metadata: order.Metadata,
			Data:     order.Data,
		})
		if err != nil {
			log.Error("can not create the payment intent", "error", err.Error())
			response.InternalServerError(c)
			return
		}

		response.Ok(c, map[string]string{"token": intent.ClientSecret})
	}
}

// CheckoutHandler returns the handler of the route creating a checkout session. It asks the order to orderFunc, builds
// the URLs where Stripe sends the user back from the web routes of the server, then answers {"token": <session ID>}
// for Stripe.js, or 500 if Stripe or the Store fails.
func (s *Service) CheckoutHandler(orderFunc OrderFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		order, ok := readOrder(c, orderFunc)
		if !ok {
			return
		}

		locale := ctxLocale.Get(c)
		routes := weburl.Of(c)

		checkout, err := s.CreateCheckout(c.Request.Context(), CheckoutRequest{
			Customer:              order.Customer,
			Items:                 order.Items,
			SuccessURL:            routes.GetURLWithParams(order.SuccessRoute, locale, order.SuccessParams),
			CancelURL:             routes.GetURLWithParams(order.CancelRoute, locale, order.CancelParams),
			RequireBillingAddress: order.RequireBillingAddress,
			Metadata:              order.Metadata,
			Data:                  order.Data,
		})
		if err != nil {
			log.Error("can not create the payment checkout", "error", err.Error())
			response.InternalServerError(c)
			return
		}

		response.Ok(c, map[string]string{"token": checkout.SessionID})
	}
}

// readOrder asks the order to orderFunc and checks it. It returns false once the request is answered: by orderFunc if
// it refuses the order, with 500 if the order can not be paid, which is an error of the application.
func readOrder(c *gin.Context, orderFunc OrderFunc) (Order, bool) {
	order, ok := orderFunc(c)
	if !ok {
		return Order{}, false
	}

	if err := order.validate(); err != nil {
		log.Error("invalid order", "error", err.Error())
		response.InternalServerError(c)
		return Order{}, false
	}

	if order.Customer.UserID == "" {
		order.Customer.UserID = ctxUser.GetID(c)
	}

	return order, true
}
