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

type Config struct {
	NetworkPublicKey   provider.NetworkPublicKeyHexed
	ProviderPrivateKey network.PrivateKeyHexed
	TZeroEndpoint      string
	ServerAddr         string
}

func main() {
	config := loadConfig()

	networkClient := initNetworkClient(config)

	shutdownFunc := startProviderServer(config, networkClient)
	defer shutdownFunc()

	// ✅ Step 1.1 is done. You successfully initialised starter template

	// TODO: Step 1.2 Share the generated public key from .env with t-0 team

	// TODO: Step 1.3 Replace publishQuotes with your own quote publishing logic

	go publishQuotes(context.Background(), networkClient)

	// TODO: Step 1.4 Verify that quotes for target currency are successfully received
	go getQuote(context.Background(), networkClient)

	waitForShutdownSignal(shutdownFunc)

	// TODO: Step 2.2 Deploy your integration and provide t-0 team with the base URL
	// TODO: Step 2.3 Test payment submission (see submit_payment.ts)
	// TODO: Step 2.5 Ask t-0 team to submit a payment to test your payOut endpoint
}

func loadConfig() Config {
	if err := godotenv.Load(".env"); err != nil {
		log.Fatalf("Failed to load .env file: %v", err)
	}

	return Config{
		NetworkPublicKey:   provider.NetworkPublicKeyHexed(os.Getenv("NETWORK_PUBLIC_KEY")),
		ProviderPrivateKey: network.PrivateKeyHexed(os.Getenv("PROVIDER_PRIVATE_KEY")),
		TZeroEndpoint:      os.Getenv("TZERO_ENDPOINT"),
		ServerAddr:         ":8080",
	}
}

func initNetworkClient(config Config) paymentconnect.NetworkServiceClient {
	networkClient, err := network.NewServiceClient(
		config.ProviderPrivateKey,
		paymentconnect.NewNetworkServiceClient,
		network.WithBaseURL(config.TZeroEndpoint),
	)
	if err != nil {
		log.Fatalf("Failed to create network service client: %v", err)
	}
	return networkClient
}

func startProviderServer(config Config, networkClient paymentconnect.NetworkServiceClient) func() {
	providerServiceHandler, err := provider.NewHttpHandler(
		config.NetworkPublicKey,
		provider.Handler(paymentconnect.NewProviderServiceHandler,
			paymentconnect.ProviderServiceHandler(handler.NewProviderServiceImplementation(networkClient))),
	)
	if err != nil {
		log.Fatalf("Failed to create provider service handler: %v", err)
	}

	shutdownFunc, err := provider.StartServer(
		providerServiceHandler,
		provider.WithAddr(config.ServerAddr),
	)
	if err != nil {
		log.Fatalf("Failed to start provider server: %v", err)
	}

	log.Printf("✅ Step 1.1: Provider server initialized on %s\n", config.ServerAddr)

	return func() {
		if err := shutdownFunc(context.Background()); err != nil {
			log.Fatalf("Failed to shutdown provider service: %v", err)
		}
	}
}

func waitForShutdownSignal(shutdownFunc func()) {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	<-ctx.Done()

	log.Println("Shutting down...")
	shutdownFunc()
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
