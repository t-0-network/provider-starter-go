package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/t-0-network/provider-sdk-go/api/tzero/v1/payment/paymentconnect"
	"github.com/t-0-network/provider-sdk-go/network"
	"github.com/t-0-network/provider-sdk-go/provider"
	"github.com/t-0-network/provider-starter-go/template/internal/handler"
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

	ctx, stop := signal.NotifyContext(context.Background(),
		os.Interrupt,
		syscall.SIGTERM)
	defer stop()

	<-ctx.Done() // blocks until signal received

	err = shutdownFunc(context.Background())
	if err != nil {
		log.Fatalf("Failed to shutdown provider service: %v", err)
	}
}
