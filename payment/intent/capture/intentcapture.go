package apppaymentintentcapture

import (
	"errors"

	appconfig "shopping-service.com/m/config"
	appcurrency "shopping-service.com/m/currency"
	apperror "shopping-service.com/m/error"
	apppaymentintent "shopping-service.com/m/payment/intent"

	"github.com/stripe/stripe-go"
	"github.com/stripe/stripe-go/paymentintent"
)

// Capture gets the intent id from c Stripe account and capture it
func Capture(id string, c appcurrency.Currency) (apppaymentintent.Intent, error) {
	if id == "" || c == nil {
		return nil, errors.New("impossible to capture the payment intent without required parameters")
	}

	sck, e := appconfig.GetStripeAPIConfigByCurrency(c.GetISO4217())
	if e != nil {
		return nil, e
	}

	stripe.Key = sck.GetSK()

	intent, e := paymentintent.Capture(id, nil)
	if e != nil {
		m, es := apperror.GetStripeErrorMessage(e)
		if es == nil {
			return nil, errors.New(m)
		}

		return nil, e
	}

	return apppaymentintent.FromStripeToAppIntent(*intent), nil
}
