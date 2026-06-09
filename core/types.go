// Package core is the chain-agnostic commerce core. It imports NO chain SDK; all chain
// specifics live behind SettlementAdapter (see adapters/<chain>).
package core

// Attestation is a chain-neutral agent decision to be written on-chain.
type Attestation struct {
	Summary string
	Metric  uint64
}

// AttestResult is returned after an attestation is recorded on-chain.
type AttestResult struct {
	TxHash string
	ID     uint32
}

// Receipt is a chain-neutral view of a previously recorded attestation.
type Receipt struct {
	Author  string
	Summary string
	Metric  uint64
	Ts      uint64
}

// PaymentReq describes an x402 payment to settle (chain-neutral).
type PaymentReq struct {
	Network string // CAIP-2, e.g. "casper:casper-test"
	PayTo   string
	Amount  string
	Asset   string
	Name    string
	Version string
}

// SettleResult is returned after an x402 settle.
type SettleResult struct {
	TxHash string
	Payer  string
}
