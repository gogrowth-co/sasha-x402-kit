package core

import "context"

// SettlementAdapter is the chain-neutral seam between the commerce core and a specific chain.
// Implementations live in adapters/<chain> and import that chain's SDK; this package does not.
// The Casper adapter lands first (flagship); an EVM adapter implements the same interface to
// prove the seam (any chain that can record an attestation satisfies it).
//
// SPINE surface is Attest — the agent's verifiable on-chain write, shipped + proven on casper-test.
// Roadmap (EXPOSE / Phase 2): ReadReceipt(id) for read-back and Settle(req) for programmatic x402
// settlement are deliberately NOT on this interface yet so it advertises only what is implemented.
type SettlementAdapter interface {
	// Attest writes an agent decision on-chain and returns the tx hash (+ record id when known).
	Attest(ctx context.Context, a Attestation) (AttestResult, error)
}
