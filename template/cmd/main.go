package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"connectrpc.com/connect"
	"github.com/joho/godotenv"
	"github.com/t-0-network/provider-sdk-go/api/tzero/v1/common"
	"github.com/t-0-network/provider-sdk-go/api/tzero/v1/payment"
	"github.com/t-0-network/provider-sdk-go/api/tzero/v1/payment/paymentconnect"
	"github.com/t-0-network/provider-sdk-go/network"
	"github.com/t-0-network/provider-sdk-go/provider"
	"github.com/t-0-network/provider-starter-go/template/internal/handler"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func main() {
	err := godotenv.Load(".env")
	networkPublicKey := provider.NetworkPublicKeyHexed(os.Getenv("NETWORK_PUBLIC_KEY"))
	providerPrivateKey := network.PrivateKeyHexed(os.Getenv("PROVIDER_PRIVATE_KEY"))
	baseUrl := os.Getenv("TZERO_ENDPOINT")

	networkClient, err := network.NewServiceClient(
		providerPrivateKey,
		paymentconnect.NewNetworkServiceClient,
		// Optional configuration for the network service client.
		network.WithBaseURL(baseUrl),
	)
	if err != nil {
		log.Fatalf("Failed to create network service client: %v", err)
	}

	// Initialize a provider service handler using your implementation of the
	// networkconnect.ProviderServiceHandler interface.
	providerServiceHandler, err := provider.NewHttpHandler(
		networkPublicKey,
		// Your provider service implementation
		provider.Handler(paymentconnect.NewProviderServiceHandler,
			paymentconnect.ProviderServiceHandler(handler.NewProviderServiceImplementation(networkClient))),
	)
	if err != nil {
		log.Fatalf("Failed to create provider service handler: %v", err)
	}

	// Start an HTTP server with the provider service handler,
	shutdownFunc, err := provider.StartServer(
		providerServiceHandler,
		provider.WithAddr(":8080"),
	)
	if err != nil {
		log.Fatalf("Failed to start provider server: %v", err)
	}

	// ✅ Step 1.1 is done. You successfully initialised starter template

	// TODO: Step 1.2 take you generated public key from .env and share it with t-0 team

	// TODO: Step 1.3 implement publishing of quotes
	go func() {
		ctx := context.Background()
		publishQuotes(ctx, networkClient)
	}()

	// TODO: Step 1.4 check that quote for target currency is successfully received
	go func() {
		ctx := context.Background()
		getQuote(ctx, networkClient)
	}()

	ctx, stop := signal.NotifyContext(context.Background(),
		os.Interrupt,
		syscall.SIGTERM)
	defer stop()

	<-ctx.Done() // blocks until signal received

	err = shutdownFunc(context.Background())
	if err != nil {
		log.Fatalf("Failed to shutdown provider service: %v", err)
	}

	// TODO: Step 2.2 deploy your integration and provide t-0 team base URL of your deployment

	// TODO: Step 2.3 check that you can submit payment by revisiting ./submit_payment.ts uncommenting following line
	// await submitPayment(networkClient)

	// TODO: Step 2.5 ask t-0 team to submit a payment which would trigger your payOut endpoin
}

func publishQuotes(ctx context.Context, networkClient paymentconnect.NetworkServiceClient) {
	// TODO: Step 1.3 replace this with fetching quotes from your systems and publishing them into t-0 Network.
	// We recommend publishing at least once per 5 seconds, but not more than once per second
	_, err := networkClient.UpdateQuote(ctx, connect.NewRequest(&payment.UpdateQuoteRequest{
		PayOut: []*payment.UpdateQuoteRequest_Quote{
			{
				Currency:      "BRL",
				QuoteType:     payment.QuoteType_QUOTE_TYPE_REALTIME, // REALTIME is only one supported right now
				PaymentMethod: common.PaymentMethodType_PAYMENT_METHOD_TYPE_CARD,
				Expiration:    timestamppb.New(time.Now().Add(30 * time.Second)), // expiration time - 30 seconds from now
				Timestamp:     timestamppb.New(time.Now()),                       // current timestamp
				Bands: []*payment.UpdateQuoteRequest_Quote_Band{ // one or more bands are allowed
					{
						ClientQuoteId: "brl-card-1000-band-1",
						MaxAmount: &common.Decimal{
							Unscaled: 1000, // maximum amount in USD, could be 1000, 5000, 10000 or 25000
							Exponent: 0,
						},
						// note that rate is always USD/XXX, so that for BRL quote should be USD/BRL
						Rate: &common.Decimal{ //rate 5.32
							Unscaled: 532,
							Exponent: -2,
						},
					},
				},
			},
		},
		// it can be either pay-in or pay-out quotes, depends on whether you want to accept incoming payments or send outgoing ones,
		// or the both.
		//PayIn: []*payment.UpdateQuoteRequest_Quote{
		//	{},
		//},
	}))
	if err != nil {
		log.Printf("Error updating quote: %s\n", err.Error()) // handle errors appropriately
		return
	}
}

func getQuote(ctx context.Context, networkClient paymentconnect.NetworkServiceClient) {
	quote, err := networkClient.GetQuote(ctx, connect.NewRequest(&payment.GetQuoteRequest{
		PayInCurrency: "BRL",
		Amount: &payment.PaymentAmount{Amount: &payment.PaymentAmount_PayInAmount{
			PayInAmount: &common.Decimal{Unscaled: 500, Exponent: 0}, // amount in BRL
		}},
		PayInMethod:    common.PaymentMethodType_PAYMENT_METHOD_TYPE_CARD,
		PayOutCurrency: "EUR",
		PayOutMethod:   common.PaymentMethodType_PAYMENT_METHOD_TYPE_WIRE,
		QuoteType:      payment.QuoteType_QUOTE_TYPE_REALTIME,
	}))
	if err != nil {
		log.Printf("Error getting quote: %s\n", err.Error()) // handle errors appropriately
		return
	} else {
		switch quote.Msg.Result.(type) {
		case *payment.GetQuoteResponse_Success_:
			log.Printf("Got success response with reason with quote id: %d \n", quote.Msg.GetSuccess().QuoteId.QuoteId)
		case *payment.GetQuoteResponse_Failure_:
			log.Printf("Got failure response with reason: %d\n", quote.Msg.GetFailure().Reason)
		default:
			log.Println("Unknown type")
		}
	}
}
