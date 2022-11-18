package apppaymentintentconfirm

import (
	"errors"
	"fmt"
	"math/rand"
	"os"
	"time"

	appconfig "shopping-service.com/m/config"
	appcurrency "shopping-service.com/m/currency"
	apperror "shopping-service.com/m/error"
	apppaymentintent "shopping-service.com/m/payment/intent"
	"solace.dev/go/messaging"
	"solace.dev/go/messaging/pkg/solace/config"
	"solace.dev/go/messaging/pkg/solace/message"
	"solace.dev/go/messaging/pkg/solace/resource"

	"github.com/stripe/stripe-go"
	"github.com/stripe/stripe-go/paymentintent"
)

// Define Topic Prefix
const TopicPrefix = "events/payment-service"

func MessageHandler(message message.InboundMessage) {
	fmt.Printf("Message Dump %s \n", message)
}

func getEnv(key, def string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return def
}

func randomString(length int) string {
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, length)
	rand.Read(b)
	return fmt.Sprintf("%x", b)[:length]
}

// Confirm gets the intent id from c Stripe account and confirm it
func Confirm(id string, c appcurrency.Currency) (apppaymentintent.Intent, error) {
	if id == "" || c == nil {
		return nil, errors.New("impossible to confirm the payment intent without required parameters")
	}

	sck, e := appconfig.GetStripeAPIConfigByCurrency(c.GetISO4217())
	if e != nil {
		return nil, e
	}

	stripe.Key = sck.GetSK()

	intent, e := paymentintent.Confirm(id, nil)
	if e != nil {
		m, es := apperror.GetStripeErrorMessage(e)
		if es == nil {
			return nil, errors.New(m)
		}

		return nil, e
	}

	// Send message to Solace broker

	// Configuration parameters
	brokerConfig := config.ServicePropertyMap{
		config.TransportLayerPropertyHost:                getEnv("TransportLayerPropertyHost", "tcps:"),
		config.ServicePropertyVPNName:                    getEnv("ServicePropertyVPNName", "brokerName"),
		config.AuthenticationPropertySchemeBasicUserName: getEnv("AuthenticationPropertySchemeBasicUserName", "solace-cloud-client"),
		config.AuthenticationPropertySchemeBasicPassword: getEnv("AuthenticationPropertySchemeBasicPassword", "password"),
	}
	messagingService, err := messaging.NewMessagingServiceBuilder().FromConfigurationProvider(brokerConfig).WithTransportSecurityStrategy(config.NewTransportSecurityStrategy().WithoutCertificateValidation()).
		Build()

	if err != nil {
		panic(err)
	}

	// Connect to the messaging serice
	if err := messagingService.Connect(); err != nil {
		panic(err)
	}

	fmt.Println("Connected to the broker? ", messagingService.IsConnected())

	//  Build a Direct Message Publisher
	directPublisher, builderErr := messagingService.CreateDirectMessagePublisherBuilder().Build()
	if builderErr != nil {
		panic(builderErr)
	}

	startErr := directPublisher.Start()
	if startErr != nil {
		panic(startErr)
	}

	fmt.Println("Direct Publisher running? ", directPublisher.IsRunning())

	//  Prepare outbound message payload and body
	messageBody := "Payment intent confirmed has id is : "
	messageBuilder := messagingService.MessageBuilder().
		WithProperty("application", "samples").
		WithProperty("language", "go")

	println("Subscribe to topic ", TopicPrefix+"/>")

	productId := randomString(5)
	paymentId := randomString(6)

	if directPublisher.IsReady() {
		message, err := messageBuilder.BuildWithStringPayload(messageBody + id)
		if err != nil {
			panic(err)
		}
		publishErr := directPublisher.Publish(message, resource.TopicOf(TopicPrefix+"/"+productId+"/"+c.GetISO4217()+"/"+"pm_card_visa/"+paymentId+"/"))
		if publishErr != nil {
			panic(publishErr)
		}
	}

	// TODO
	// Find way to shutdown the go routine
	// e.g use another channel, BOOl..etc
	// TODO

	return apppaymentintent.FromStripeToAppIntent(*intent), nil
}
