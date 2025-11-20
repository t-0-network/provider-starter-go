package internal

import (
	"context"
	"log"
	"time"

	"connectrpc.com/connect"
	"github.com/t-0-network/provider-sdk-go/api/tzero/v1/common"
	"github.com/t-0-network/provider-sdk-go/api/tzero/v1/payment"
	"github.com/t-0-network/provider-sdk-go/api/tzero/v1/payment/paymentconnect"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func PublishQuotes(ctx context.Context, networkClient paymentconnect.NetworkServiceClient) {
	// TODO: Step 1.3 replace this with fetching quotes from your systems and publishing them into t-0 Network.
	// We recommend publishing at least once per 5 seconds, but not more than once per second

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			_, err := networkClient.UpdateQuote(ctx, connect.NewRequest(&payment.UpdateQuoteRequest{
				PayOut: []*payment.UpdateQuoteRequest_Quote{
					{
						Currency:      "EUR",
						QuoteType:     payment.QuoteType_QUOTE_TYPE_REALTIME, // REALTIME is only one supported right now
						PaymentMethod: common.PaymentMethodType_PAYMENT_METHOD_TYPE_CARD,
						Expiration:    timestamppb.New(time.Now().Add(30 * time.Second)), // expiration time - 30 seconds from now
						Timestamp:     timestamppb.New(time.Now()),                       // current timestamp
						Bands: []*payment.UpdateQuoteRequest_Quote_Band{ // one or more bands are allowed
							{
								ClientQuoteId: "eur-card-1000-band-1",
								MaxAmount: &common.Decimal{
									Unscaled: 1000, // maximum amount in USD, could be 1000, 5000, 10000 or 25000
									Exponent: 0,
								},
								// note that rate is always USD/XXX, so that for BRL quote should be USD/BRL
								Rate: &common.Decimal{ //rate 0.86
									Unscaled: 86,
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
	}
}
