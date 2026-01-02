package handler

import (
	"context"

	"connectrpc.com/connect"
	"github.com/t-0-network/provider-sdk-go/api/tzero/v1/payment"
	"github.com/t-0-network/provider-sdk-go/api/tzero/v1/payment/paymentconnect"
)

type ProviderServiceImplementation struct {
	networkClient paymentconnect.NetworkServiceClient
}

func NewProviderServiceImplementation(networkClient paymentconnect.NetworkServiceClient) *ProviderServiceImplementation {
	return &ProviderServiceImplementation{
		networkClient: networkClient,
	}
}

/*
  Please refer to docs, proto definition comments or source code comments to understand purpose of functions
*/

var _ paymentconnect.ProviderServiceHandler = (*ProviderServiceImplementation)(nil)

// TODO: Step 2.1 implement how you handle updates of payment initiated by you
func (s *ProviderServiceImplementation) UpdatePayment(
	ctx context.Context, req *connect.Request[payment.UpdatePaymentRequest],
) (*connect.Response[payment.UpdatePaymentResponse], error) {
	return connect.NewResponse(&payment.UpdatePaymentResponse{}), nil
}

// TODO: Step 2.4 implement how you do payouts (payments initiated by your counterparts)
func (s *ProviderServiceImplementation) PayOut(ctx context.Context, req *connect.Request[payment.PayoutRequest],
) (*connect.Response[payment.PayoutResponse], error) {
	return connect.NewResponse(&payment.PayoutResponse{}), nil
}

//TODO: confirmPayout should be called when you system notifies that payout has been made successfully

func (s *ProviderServiceImplementation) UpdateLimit(
	ctx context.Context, req *connect.Request[payment.UpdateLimitRequest],
) (*connect.Response[payment.UpdateLimitResponse], error) {
	// TODO: optionally implement handling of the notifications about updates on your limits and limits usage
	return connect.NewResponse(&payment.UpdateLimitResponse{}), nil
}

func (s *ProviderServiceImplementation) AppendLedgerEntries(
	ctx context.Context, req *connect.Request[payment.AppendLedgerEntriesRequest],
) (*connect.Response[payment.AppendLedgerEntriesResponse], error) {
	// TODO: optionally implement handling of the notifications about new ledger transactions and new ledger entries
	return connect.NewResponse(&payment.AppendLedgerEntriesResponse{}), nil
}

// TODO: Step 2.5 when the payment goes through the Manual AML Check on the pay-out provider side, the provider submitted the payment will have a last look to approve final quote
func (s *ProviderServiceImplementation) ApprovePaymentQuotes(ctx context.Context, c *connect.Request[payment.ApprovePaymentQuoteRequest]) (*connect.Response[payment.ApprovePaymentQuoteResponse], error) {
	//TODO: check the pay-out quote and decide if it's ok for you to proceed with the payment
	return connect.NewResponse(&payment.ApprovePaymentQuoteResponse{}), nil
}
