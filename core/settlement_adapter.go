package core

import "context"

// SettlementAdapter is the chain-neutral seam between the commerce core and a specific chain.
// Implementations live in adapters/<chain> and import that chain's SDK; this package does not.
// The Casper adapter lands first (flagship); an EVM adapter proves the seam later.
type SettlementAdapter interface {
	// Attest writes an agent decision on-chain and returns the tx hash (+ record id when known).
	Attest(ctx context.Context, a Attestation) (AttestResult, error)

	// ReadReceipt reads back a recorded attestation by id.
	ReadReceipt(ctx context.Context, id uint32) (Receipt, error)

	// Settle submits an x402 payment on-chain and returns the settle tx hash.
	Settle(ctx context.Context, req PaymentReq) (SettleResult, error)
}
