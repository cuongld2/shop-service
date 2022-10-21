//go:build unit
// +build unit

package apppaymentsource_test

import (
	"testing"

	apppaymentsource "shopping-service.com/m/payment/source"
)

func TestNew(t *testing.T) {
	r := "card_XXX"
	got := apppaymentsource.New(r)

	if got.GetGatewayReference() != r {
		t.Errorf("New payment source reference is incorrect, got %v want %v", got.GetGatewayReference(), r)
	}
}
