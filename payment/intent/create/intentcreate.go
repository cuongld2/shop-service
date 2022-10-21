package apppaymentintentcreate

import (
	"errors"

	appamount "shopping-service.com/m/amount"
	appconfig "shopping-service.com/m/config"
	appcustomer "shopping-service.com/m/customer"
	apperror "shopping-service.com/m/error"
	apppaymentintent "shopping-service.com/m/payment/intent"
	apppaymentsource "shopping-service.com/m/payment/source"

	"github.com/stripe/stripe-go"
	"github.com/stripe/stripe-go/paymentintent"
)

// Create creates an intent in Stripe and returns it as an instance of Intent
func Create(a appamount.Amount, p apppaymentsource.Source, c appcustomer.Customer) (apppaymentintent.Intent, error) {
	if a == nil || p == nil {
		return nil, errors.New("impossible to create a payment intent without required parameters")
	}

	sck, e := appconfig.GetStripeAPIConfigByCurrency(a.GetCurrency().GetISO4217())
	if e != nil {
		return nil, e
	}

	stripe.Key = sck.GetSK()

	ic := &stripe.PaymentIntentParams{
		Amount:             stripe.Int64(int64(a.GetAmount())),
		Currency:           stripe.String(a.GetCurrency().GetISO4217()),
		PaymentMethod:      stripe.String(p.GetGatewayReference()),
		SetupFutureUsage:   stripe.String("off_session"),
		ConfirmationMethod: stripe.String("manual"),
		CaptureMethod:      stripe.String("manual"),
	}

	if c != nil {
		ic.Customer = stripe.String(c.GetGatewayReference())
		ic.SavePaymentMethod = stripe.Bool(true)
	}

	intent, e := paymentintent.New(ic)
	if e != nil {
		m, es := apperror.GetStripeErrorMessage(e)
		if es == nil {
			return nil, errors.New(m)
		}

		return nil, e
	}

	return apppaymentintent.FromStripeToAppIntent(*intent), nil
}
